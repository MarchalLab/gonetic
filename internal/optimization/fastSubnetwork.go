package optimization

import (
	"math/rand"
)

// *fastSubnetwork implements subnetwork, using random selections for expansions
// ! Should only be used if the total amount of paths is a couple orders of magnitude higher than the usual subnetwork size
type fastSubnetwork struct {
	*genericSubnetwork
}

func newFastSubnetwork(opt *NSGAOptimization) *fastSubnetwork {
	return &fastSubnetwork{
		genericSubnetwork: newGenericSubnetwork(opt),
	}
}

// expansion adds a random path to the subnetwork as nsga mutation operator
func (network *fastSubnetwork) expansion() {
	// set number of tries to generate new expansion to amount of paths
	// this should normally never be a concern as the total amount of paths should be several magnitudes larger than the size of a subnetwork in our application
	// for other applications where this is not the case, the original nsga subnetwork can be used
	maxTries := network.opt.pathRepositories.NumberOfPaths()

	for tries := 0; tries <= maxTries; tries++ {
		// randomly select a repository
		repo, repoIndex := network.opt.pathRepositories.RandomRepo()
		// randomly select a sample
		_, sampleIdx := repo.RandomSample()
		// randomly select a path
		_, samplePathIdx := repo.RandomPath(sampleIdx)
		// create pathID
		pathID := NewPathID(repoIndex, sampleIdx, samplePathIdx)
		// check if any new interactions
		interactions := network.opt.pathRepositories.pathInteractionSetFromId(pathID)
		for interactionID := range interactions {
			if !network.interactionSet.Has(interactionID) {
				// a suitable path for expansion has been found, perform the expansion and exit
				network.expandWithPath(pathID)
				return
			}
		}
	}
	// if no suitable expansion has been found, then return without expanding
	return
}

// reduction removes a random path from the subnetwork as nsga mutation operator
func (network *fastSubnetwork) reduction() {

	// never create empty subnetwork
	if len(network.selectedPaths) == 1 {
		return
	}

	// get set of keys from selectedPaths
	pathIds := make([]PathID, len(network.selectedPaths))
	j := 0
	for i := range network.selectedPaths {
		pathIds[j] = i
		j++
	}
	// select random path
	p := pathIds[rand.Intn(len(pathIds))]
	// delete interactions of this path
	for i := range network.selectedPaths[p] {
		network.interactionSet.Delete(i)
	}
	// delete path from network
	delete(network.selectedPaths, p)
}
