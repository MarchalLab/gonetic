package types

import (
	"fmt"
	"strconv"
)

type TranslationMap[Name ~string] map[Name]Name

// IDMap is a map that stores the mapping between int-like IDs and string-like names
type IDMap[ID ~uint64, Name ~string] struct {
	maxID    ID
	idToName map[ID]Name
	nameToID map[Name]ID
}

// NewIDMap creates a new IDMap
func NewIDMap[ID ~uint64, Name ~string]() *IDMap[ID, Name] {
	return &IDMap[ID, Name]{
		maxID:    1,
		idToName: make(map[ID]Name),
		nameToID: make(map[Name]ID),
	}
}

func (im *IDMap[ID, Name]) NameToID() map[Name]ID {
	return im.nameToID
}

func (im *IDMap[ID, Name]) IdToName() map[ID]Name {
	return im.idToName
}

// ParseID converts a string representation of an ID to a numerical ID
func ParseID[ID ~uint64](idString string) ID {
	id, err := strconv.Atoi(idString)
	if err != nil {
		panic(fmt.Sprintf("id %s is not a number: %s", idString, err))
	}
	return ID(id)
}

// SetName adds a name to the map and returns its ID
func (im *IDMap[ID, Name]) SetName(nameString string) ID {
	name := Name(nameString)
	// check if the name is already in the map
	if knownID, knownName := im.nameToID[name]; knownName {
		return knownID
	}
	// add the name to the maps with a sequential ID
	return im.SetNameWithID(nameString, im.maxID)
}

// SetNameWithID adds a name to the map and returns its ID
func (im *IDMap[ID, Name]) SetNameWithID(nameString string, id ID) ID {
	name := Name(nameString)
	// check if the name has already been assigned an ID
	if knownID, knownName := im.nameToID[name]; knownName {
		if knownID != id {
			panic(fmt.Sprintf("SetNameWithID: name %s id %d is already known with ID %d", nameString, id, knownID))
		}
		return knownID
	}
	// check if the ID is already in use
	if knownName, knownID := im.idToName[id]; knownID {
		panic(fmt.Sprintf("SetNameWithID: name %s id %d is already known with name %s", nameString, id, knownName))
	}
	// add the name to the maps
	im.maxID = max(im.maxID, id+1)
	im.idToName[id] = name
	im.nameToID[name] = id
	return id
}

// GetIDFromName returns the ID of the given name
func (im *IDMap[ID, Name]) GetIDFromName(name Name) ID {
	if id, knownName := im.nameToID[name]; knownName {
		return id
	}
	panic(fmt.Sprintf("GetIDFromName: name %s is not known", name))
}

// GetIDFromString converts strings to IDs
//   - if the string is a known name, then it returns the ID of the given name
//   - otherwise, it parses the string as an integer and returns it as a ID
func (im *IDMap[ID, Name]) GetIDFromString(name string) ID {
	if id, knownName := im.nameToID[Name(name)]; knownName {
		return id
	}
	return ParseID[ID](name)
}

// GetNameFromID returns the name of the given ID
func (im *IDMap[ID, Name]) GetNameFromID(id ID) Name {
	if name, knownName := im.idToName[id]; knownName {
		return name
	}
	panic(fmt.Sprintf("GetNameFromID: id %d is not known", id))
}

// GetMappedName returns the translation of the given name
func (im *IDMap[ID, Name]) GetMappedName(id ID, mapping TranslationMap[Name]) Name {
	name := im.GetNameFromID(id)
	if mapped, ok := mapping[name]; ok {
		return mapped
	}
	// no mapping found, return the original name
	return name
}

type fileWriter interface {
	WriteLinesToNewFile(string, ...[]string) error
}

// WriteNameMap writes the name map to the given file
func (im *IDMap[ID, Name]) WriteNameMap(fw fileWriter, filename string) error {
	lines := make([]string, 0, len(im.idToName))
	for id := range im.maxID {
		if _, knownName := im.idToName[id]; !knownName {
			continue
		}
		lines = append(lines, fmt.Sprintf(
			"%d\t%s",
			id,
			im.idToName[id],
		))
	}
	return fw.WriteLinesToNewFile(filename, lines)
}
