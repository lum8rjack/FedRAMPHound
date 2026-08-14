package main

import (
	"fmt"
	"regexp"
	"strings"
)

const (
	sourceKind = "FedRAMP"

	kindAgency   = "FedRAMP_Agency"
	kindAssessor = "FedRAMP_Assessor"
	kindProduct  = "FedRAMP_Product"
	kindProvider = "FedRAMP_Provider"

	edgeContains     = "FedRAMP_Contains"
	edgeAuthorizes   = "FedRAMP_Authorizes"
	edgeReuses       = "FedRAMP_Reuses"
	edgeInProcess    = "FedRAMP_InProcess"
	edgeAssesses     = "FedRAMP_Assesses"
	edgeProvides     = "FedRAMP_Provides"
	edgeLeverages    = "FedRAMP_Leverages"
	edgePartnersWith = "FedRAMP_PartnersWith"
)

var nonAlphaNum = regexp.MustCompile(`[^a-z0-9]+`)

func normalizeName(s string) string {
	return nonAlphaNum.ReplaceAllString(strings.ToLower(strings.TrimSpace(s)), "")
}

func cleanName(s string) string {
	return strings.TrimSpace(s)
}

func isPlaceholderAssessor(name string) bool {
	n := strings.ToLower(cleanName(name))
	if n == "" || n == "n/a" || n == "none" || n == "unknown" {
		return true
	}
	if strings.Contains(n, "no assessor") || strings.Contains(n, "not yet assigned") {
		return true
	}
	if strings.Contains(n, "self-attestation") || strings.Contains(n, "placeholder") {
		return true
	}
	return false
}

func providerID(name string) string {
	return "csp:" + cleanName(name)
}

func commonPrefixLen(a, b string) int {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	i := 0
	for i < n && a[i] == b[i] {
		i++
	}
	return i
}

type edgeKey struct {
	kind  string
	start string
	end   string
}

type graphBuilder struct {
	graph *OpenGraph

	agencyByID     map[string]Agency
	agencyByName   map[string]string // parent or sub display name -> node id
	parentAgencyID map[string]string // parent org name -> parent-level node id
	assessorByID   map[string]Assessor
	assessorByName map[string]string // exact/normalized name -> assessor id
	productIDs     map[string]struct{}
	providerIDs    map[string]struct{}
	seenEdges      map[edgeKey]struct{}
}

func BuildOpenGraph(data *FedRAMPData) *OpenGraph {
	b := &graphBuilder{
		graph:          NewOpenGraph(sourceKind),
		agencyByID:     make(map[string]Agency),
		agencyByName:   make(map[string]string),
		parentAgencyID: make(map[string]string),
		assessorByID:   make(map[string]Assessor),
		assessorByName: make(map[string]string),
		productIDs:     make(map[string]struct{}),
		providerIDs:    make(map[string]struct{}),
		seenEdges:      make(map[edgeKey]struct{}),
	}

	b.addAgencies(data.Agencies)
	b.addAssessors(data.Assessors)
	b.addProducts(data.Products)
	b.addAgencyProductEdges(data.Agencies)
	b.addProductRelationshipEdges(data.Products)
	b.addAssessorClientEdges(data.Assessors)
	b.suppressRedundantParentProductEdges()

	return b.graph
}

// suppressRedundantParentProductEdges drops parent→product links when a sub-agency
// under that parent already has the same relationship, so the graph shows
// Parent -Contains-> Sub -Authorizes/Reuses-> Product instead of two parallel edges.
func (b *graphBuilder) suppressRedundantParentProductEdges() {
	children := map[string][]string{}
	for _, e := range b.graph.Graph.Edges {
		if e.Kind == edgeContains {
			children[e.Start.Value] = append(children[e.Start.Value], e.End.Value)
		}
	}

	type rel struct {
		kind, agency, product string
	}
	has := map[rel]bool{}
	for _, e := range b.graph.Graph.Edges {
		switch e.Kind {
		case edgeAuthorizes, edgeReuses, edgeInProcess:
			has[rel{e.Kind, e.Start.Value, e.End.Value}] = true
		}
	}

	filtered := make([]*OpenGraphEdge, 0, len(b.graph.Graph.Edges))
	for _, e := range b.graph.Graph.Edges {
		switch e.Kind {
		case edgeAuthorizes, edgeReuses, edgeInProcess:
			skip := false
			for _, child := range children[e.Start.Value] {
				if has[rel{e.Kind, child, e.End.Value}] {
					skip = true
					break
				}
			}
			if skip {
				continue
			}
		}
		filtered = append(filtered, e)
	}
	b.graph.Graph.Edges = filtered
}

func (b *graphBuilder) addEdge(kind, start, end string, props map[string]any) {
	if start == "" || end == "" || start == end {
		return
	}
	key := edgeKey{kind: kind, start: start, end: end}
	if _, ok := b.seenEdges[key]; ok {
		return
	}
	b.seenEdges[key] = struct{}{}
	b.graph.AddEdge(kind, start, end, props)
}

func (b *graphBuilder) ensureProvider(name string) string {
	name = cleanName(name)
	if name == "" || strings.EqualFold(name, "n/a") {
		return ""
	}
	id := providerID(name)
	if _, ok := b.providerIDs[id]; ok {
		return id
	}
	b.providerIDs[id] = struct{}{}
	b.graph.AddNode(id, []string{kindProvider}, map[string]any{
		"name":        strings.ToUpper(name),
		"displayname": name,
	})
	return id
}

func looksLikeAgencyID(s string) bool {
	// Marketplace agency IDs look like "22-008" or "22-008-02".
	if len(s) < 4 || s[2] != '-' {
		return false
	}
	for i := 0; i < 2; i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

func (b *graphBuilder) resolveAgencyRef(ref string) string {
	ref = cleanName(ref)
	if ref == "" || strings.EqualFold(ref, "n/a") {
		return ""
	}
	if _, ok := b.agencyByID[ref]; ok {
		return ref
	}
	if looksLikeAgencyID(ref) {
		return ""
	}
	if id, ok := b.agencyByName[ref]; ok {
		return id
	}
	if id, ok := b.parentAgencyID[ref]; ok {
		return id
	}
	// Only agencies from the marketplace dump become nodes.
	return ""
}

func (b *graphBuilder) resolveAssessor(name string) string {
	name = cleanName(name)
	if isPlaceholderAssessor(name) {
		return ""
	}
	if id, ok := b.assessorByName[name]; ok {
		return id
	}
	norm := normalizeName(name)
	if id, ok := b.assessorByName[norm]; ok {
		return id
	}
	// Map older/short product labels onto the current assessors-file name only.
	// Never invent assessor nodes from product text.
	return b.fuzzyAssessorMatch(norm)
}

func (b *graphBuilder) fuzzyAssessorMatch(normQuery string) string {
	if len(normQuery) < 5 {
		return ""
	}

	bestID := ""
	bestScore := 0
	ambiguous := false

	for id, a := range b.assessorByID {
		an := normalizeName(a.Name)
		score := 0
		switch {
		case strings.HasPrefix(an, normQuery) || strings.HasPrefix(normQuery, an):
			score = len(normQuery)
			if len(an) < score {
				score = len(an)
			}
		default:
			if n := commonPrefixLen(an, normQuery); n >= 8 {
				score = n
			}
		}
		if score > bestScore {
			bestScore = score
			bestID = id
			ambiguous = false
		} else if score > 0 && score == bestScore && id != bestID {
			ambiguous = true
		}
	}

	if ambiguous || bestID == "" {
		return ""
	}
	return bestID
}

func (b *graphBuilder) addAgencies(agencies []Agency) {
	for _, a := range agencies {
		b.agencyByID[a.ID] = a
		display := agencyDisplayName(a)
		props := map[string]any{
			"name":          strings.ToUpper(display),
			"displayname":   display,
			"parent":        a.Parent,
			"authorization": a.Authorization,
			"reuse":         a.Reuse,
			"website":       a.Website,
		}
		if a.Sub != nil && cleanName(*a.Sub) != "" {
			props["sub"] = *a.Sub
			props["is_subagency"] = true
		} else {
			props["is_subagency"] = false
			b.parentAgencyID[a.Parent] = a.ID
		}
		if email := asString(a.Email); email != "" && email != "null" {
			props["email"] = email
		}

		b.graph.AddNode(a.ID, []string{kindAgency}, props)
		b.agencyByName[display] = a.ID
		if a.Sub == nil || cleanName(*a.Sub) == "" {
			b.agencyByName[a.Parent] = a.ID
		}
	}

	for _, a := range agencies {
		if a.Sub == nil || cleanName(*a.Sub) == "" {
			continue
		}
		parentID := b.parentAgencyID[a.Parent]
		if parentID == "" {
			continue
		}
		b.addEdge(edgeContains, parentID, a.ID, nil)
	}
}

func (b *graphBuilder) addAssessors(assessors []Assessor) {
	for _, a := range assessors {
		b.assessorByID[a.ID] = a
		b.assessorByName[a.Name] = a.ID
		b.assessorByName[normalizeName(a.Name)] = a.ID

		props := map[string]any{
			"name":                          strings.ToUpper(a.Name),
			"displayname":                   a.Name,
			"products_assessing":            a.ProductsAssessing,
			"poc":                           a.POC,
			"email":                         a.Email,
			"address":                       a.Address,
			"desc":                          a.Desc,
			"highest_impact_level":          a.HighestImpactLevel,
			"highest_impact_level_number":   a.HighestImpactLevelNumber,
			"clients_in_process":            a.ClientsInProcess,
			"has_20x_assessed":              a.Has20XAssessed,
			"frameworks":                    a.Frameworks,
			"accredited_since":              formatTime(a.AccreditedSince),
			"founded":                       formatTime(a.Founded),
		}
		if a.Logo != "" {
			props["logo"] = a.Logo
		}
		if a.Services != "" && a.Services != "undefined" {
			props["services"] = a.Services
		}

		b.graph.AddNode(a.ID, []string{kindAssessor}, props)
	}
}

func (b *graphBuilder) addProducts(products []Product) {
	for _, p := range products {
		b.productIDs[p.ID] = struct{}{}

		display := p.CSO
		if display == "" {
			display = p.ID
		}
		props := map[string]any{
			"name":                 strings.ToUpper(display),
			"displayname":          display,
			"csp":                  p.CSP,
			"cso":                  p.CSO,
			"status":               p.Status,
			"phase":                p.Phase,
			"under_cap":            p.UnderCap,
			"authorization":        p.Authorization,
			"reuse":                p.Reuse,
			"cert_path":            p.CertPath,
			"cert_type":            p.CertType,
			"impact_level":         p.ImpactLevel,
			"impact_level_number":  p.ImpactLevelNumber,
			"deployment_model":     p.DeploymentModel,
			"service_model":        p.ServiceModel,
			"independent_assessor": p.IndependentAssessor,
			"website":              p.Website,
			"sales_email":          p.SalesEmail,
			"security_email":       p.SecurityEmail,
			"small_business":       p.SmallBusiness,
			"service_desc":         p.ServiceDesc,
			"service_acronym":      p.ServiceAcronym,
			"cert_date":            formatTime(p.CertDate),
			"status_date":          formatTime(p.StatusDate),
			"annual_assessment":    formatTime(p.AnnualAssessment),
		}
		if p.PartneringAgency != nil && cleanName(*p.PartneringAgency) != "" {
			props["partnering_agency"] = *p.PartneringAgency
		}
		if p.Logo != "" {
			props["logo"] = p.Logo
		}
		if uei := asString(p.UEI); uei != "" && uei != "null" {
			props["uei"] = uei
		}

		b.graph.AddNode(p.ID, []string{kindProduct}, props)

		if provider := b.ensureProvider(p.CSP); provider != "" {
			b.addEdge(edgeProvides, provider, p.ID, nil)
		}
	}
}

func (b *graphBuilder) addAgencyProductEdges(agencies []Agency) {
	for _, a := range agencies {
		for _, ref := range a.Auths {
			if _, ok := b.productIDs[ref.ID]; !ok {
				continue
			}
			b.addEdge(edgeAuthorizes, a.ID, ref.ID, map[string]any{
				"status":       ref.Status,
				"impact_level": ref.ImpactLevel,
			})
		}
		for _, ref := range a.Reuses {
			if _, ok := b.productIDs[ref.ID]; !ok {
				continue
			}
			b.addEdge(edgeReuses, a.ID, ref.ID, map[string]any{
				"status":       ref.Status,
				"impact_level": ref.ImpactLevel,
			})
		}
		for _, ref := range a.Procs {
			if _, ok := b.productIDs[ref.ID]; !ok {
				continue
			}
			b.addEdge(edgeInProcess, a.ID, ref.ID, map[string]any{
				"status":       ref.Status,
				"impact_level": ref.ImpactLevel,
			})
		}
	}
}

func (b *graphBuilder) addProductRelationshipEdges(products []Product) {
	for _, p := range products {
		for _, agencyName := range p.AgencyAuthorizations {
			agencyID := b.resolveAgencyRef(agencyName)
			if agencyID == "" {
				continue
			}
			b.addEdge(edgeAuthorizes, agencyID, p.ID, nil)
		}
		for _, agencyName := range p.AgencyReuse {
			agencyID := b.resolveAgencyRef(agencyName)
			if agencyID == "" {
				continue
			}
			b.addEdge(edgeReuses, agencyID, p.ID, nil)
		}

		if assessorID := b.resolveAssessor(p.IndependentAssessor); assessorID != "" {
			b.addEdge(edgeAssesses, assessorID, p.ID, nil)
		}

		if p.PartneringAgency != nil {
			agencyID := b.resolveAgencyRef(*p.PartneringAgency)
			if agencyID != "" {
				b.addEdge(edgePartnersWith, p.ID, agencyID, nil)
			}
		}

		for _, leveraged := range p.LeveragedSystems {
			if leveraged.ID == "" {
				continue
			}
			if _, ok := b.productIDs[leveraged.ID]; !ok {
				// Referenced leveraged offering may be the product itself or missing; skip unknowns.
				continue
			}
			b.addEdge(edgeLeverages, p.ID, leveraged.ID, map[string]any{
				"status":       leveraged.Status,
				"impact_level": leveraged.ImpactLevel,
			})
		}
	}
}

func (b *graphBuilder) addAssessorClientEdges(assessors []Assessor) {
	for _, a := range assessors {
		for _, client := range a.Clients {
			if _, ok := b.productIDs[client.ID]; !ok {
				continue
			}
			b.addEdge(edgeAssesses, a.ID, client.ID, map[string]any{
				"status":       client.Status,
				"impact_level": client.ImpactLevel,
			})
		}
	}
}

func summarizeGraph(g *OpenGraph) string {
	kindCounts := map[string]int{}
	for _, n := range g.Graph.Nodes {
		if len(n.Kinds) > 0 {
			kindCounts[n.Kinds[0]]++
		}
	}
	edgeCounts := map[string]int{}
	for _, e := range g.Graph.Edges {
		edgeCounts[e.Kind]++
	}
	return fmt.Sprintf("nodes=%d edges=%d kinds=%v edgeKinds=%v",
		len(g.Graph.Nodes), len(g.Graph.Edges), kindCounts, edgeCounts)
}
