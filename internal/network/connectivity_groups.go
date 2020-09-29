package network

import (
	"encoding/json"
	"net"
	"sort"

	"github.com/go-openapi/strfmt"
	"github.com/openshift/assisted-service/models"
)

type connectivityKey struct {
	first, second strfmt.UUID
}

type connectivityValue struct {
	first2second, second2first bool
}

// Map for indicating if there is mutual connectivity between 2 hosts
type connectivityMap map[connectivityKey]*connectivityValue

func makeKey(from, to strfmt.UUID) connectivityKey {
	if from < to {
		return connectivityKey{first: from, second: to}
	} else {
		return connectivityKey{first: to, second: from}
	}
}

func (c connectivityMap) add(from, to strfmt.UUID, connected bool) {
	key := makeKey(from, to)
	value, ok := c[key]
	if !ok {
		value = &connectivityValue{}
		c[key] = value
	}
	if from == key.first {
		value.first2second = connected
	} else {
		value.second2first = connected
	}
}

func (c connectivityMap) isConnected(from, to strfmt.UUID) bool {
	value, ok := c[makeKey(from, to)]
	return ok && value.first2second && value.second2first
}

type connectivitySet map[strfmt.UUID]bool

func (c connectivitySet) add(item strfmt.UUID) {
	c[item] = true
}

func (c connectivitySet) intersect(other connectivitySet) connectivitySet {
	ret := make(connectivitySet)
	var first, second connectivitySet
	if len(c) < len(other) {
		first = c
		second = other
	} else {
		first = other
		second = c
	}
	for k := range first {
		_, ok := second[k]
		if ok {
			ret[k] = true
		}
	}
	return ret
}

func (c connectivitySet) containsElement(id strfmt.UUID) bool {
	_, ok := c[id]
	return ok
}

func (c connectivitySet) equals(other connectivitySet) bool {
	if len(c) != len(other) {
		return false
	}
	intersection := c.intersect(other)
	return len(intersection) == len(c)
}

func (c connectivitySet) isSupersetOf(other connectivitySet) bool {
	intersection := c.intersect(other)
	return len(intersection) == len(other)
}

type groupId []strfmt.UUID

func (g groupId) isLess(other groupId) bool {
	if len(g) != len(other) {
		return len(g) < len(other)
	}
	index := 0
	for ; index != len(g) && g[index] == other[index]; index++ {
	}
	return index < len(g) && g[index] < other[index]
}

func (c connectivitySet) id() groupId {
	ret := make(groupId, 0)
	for k := range c {
		ret = append(ret, k)
	}
	sort.Slice(ret, func(i, j int) bool {
		return ret[i] < ret[j]
	})
	return ret
}

func (c connectivitySet) toList() []strfmt.UUID {
	ret := make([]strfmt.UUID, 0)
	for k := range c {
		ret = append(ret, k)
	}
	return ret
}

type connectivityGroup struct {
	set   connectivitySet
	count int
}

// Unique list of connectivity group elements.  The uniqueness is by the set elements
type connectivityGroupList struct {
	groups []*connectivityGroup
}

func (c *connectivityGroupList) containsSet(cs connectivitySet) bool {
	for _, group := range c.groups {
		if group.set.equals(cs) {
			return true
		}
	}
	return false
}

func (c *connectivityGroupList) addSet(cs connectivitySet) {
	// Verify uniqueness
	if !c.containsSet(cs) {
		c.groups = append(c.groups, &connectivityGroup{
			set:   cs,
			count: 0,
		})
	}
}

type groupCandidate struct {
	set connectivitySet
	me  strfmt.UUID
}

// Create sorted list of connectivity sets.  The sort is by set size (descending)
func createConnectivityGroups(groupCandidates []groupCandidate) []connectivitySet {
	var groupList connectivityGroupList

	// First iteration - gather the sets
	for _, candidate := range groupCandidates {

		// All sets pending for insertion
		pendingSets := make([]connectivitySet, 0)

		// Add the set of the current candidate to the pending list
		pendingSets = append(pendingSets, candidate.set)
		for _, group := range groupList.groups {

			// Intersect the set of the current candidate with each member of the groupList.  The result is added to the
			// Pending sets
			set := candidate.set.intersect(group.set)
			if len(set) >= 3 {
				pendingSets = append(pendingSets, set)
			}
		}

		// Add the sets in the pending list to the groupList
		for _, set := range pendingSets {
			groupList.addSet(set)
		}
	}

	// Second iteration - Per groupList element. count the number of candidates that are part of that element.
	// Since every candidate represents a host, if the number of participants == the set size, then there is a full
	// mesh connectivity between the members
	for _, candidate := range groupCandidates {
		for _, cs := range groupList.groups {
			if candidate.set.isSupersetOf(cs.set) && cs.set.containsElement(candidate.me) {
				cs.count++
			}
		}
	}
	ret := make([]connectivitySet, 0)
	for _, r := range groupList.groups {
		// Add only sets with full mesh connectivity
		if len(r.set) == r.count {
			ret = append(ret, r.set)
		}
	}

	// Sort by set size descending, which means the largest group first.
	sort.Slice(ret, func(i, j int) bool {
		if len(ret[i]) != len(ret[j]) {
			return len(ret[i]) > len(ret[j])
		}

		// If the sizes are equal then compare the contained elements of each set
		return ret[i].id().isLess(ret[j].id())
	})
	return ret
}

/*
 * Create group candidate for a specific host.  The group candidate contains a set with all the hosts it has
 * connectivity to.
 */
func createHostGroupCandidate(host *models.Host, hosts []*models.Host, cMap connectivityMap) groupCandidate {
	set := make(connectivitySet)
	set.add(*host.ID)
	for _, h := range hosts {
		if cMap.isConnected(*host.ID, *h.ID) {
			set.add(*h.ID)
		}
	}
	return groupCandidate{
		set: set,
		me:  *host.ID,
	}
}

/*
 * Create connectivity map from host list.  It is the information if a host has connectivity to other host on specific
 * CIDR (network)
 */
func createMachineCidrConnectivityMap(cidr string, hosts []*models.Host) (connectivityMap, error) {
	ret := make(connectivityMap)
	_, parsedCidr, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}
	for _, h := range hosts {
		if h.Connectivity == "" {
			continue
		}
		var connectivityReport models.ConnectivityReport
		err = json.Unmarshal([]byte(h.Connectivity), &connectivityReport)
		if err != nil {
			return nil, err
		}
		for _, r := range connectivityReport.RemoteHosts {
			for _, l2 := range r.L2Connectivity {
				ip := net.ParseIP(l2.OutgoingIPAddress)
				if ip != nil && parsedCidr.Contains(ip) && l2.Successful {
					ret.add(*h.ID, r.HostID, true)
					break
				}
			}
		}
	}
	return ret, nil
}

/*
 * Crate majority for a cidr.  A majority group is a the largest group of hosts in a cluster that all of them have full mesh
 * to the other group members.
 * It is done by taking a sorted connectivity group list according to the group size, and from this group take the
 * largest one
 */
func CreateMajorityGroup(cidr string, hosts []*models.Host) ([]strfmt.UUID, error) {
	cMap, err := createMachineCidrConnectivityMap(cidr, hosts)
	if err != nil {
		return nil, err
	}
	candidates := make([]groupCandidate, 0)
	for _, h := range hosts {
		candidate := createHostGroupCandidate(h, hosts, cMap)
		if len(candidate.set) >= 3 {
			candidates = append(candidates, createHostGroupCandidate(h, hosts, cMap))
		}
	}
	groups := createConnectivityGroups(candidates)
	if len(groups) > 0 {
		return groups[0].toList(), nil
	}
	return make([]strfmt.UUID, 0), nil
}
