package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/averseabfun/logger"
)

var AclangVersion = VersionFromSemVer("v0.1.0")
var SupportedRuntimes = SupportedVersion{BaseVersion: AclangVersion, ToVersion: AclangVersion, OrHigher: true}

// Note to self: DO NOT USE any codes between 20-7F inclusive.
const (
	internalStartingRoom       = "\x01\x00"
	internalBeginLookup        = "\x01\x01"
	internalEndLookup          = "\x01\x02"
	internalEndCurLookup       = "\x01\x03"
	internalAdventureName      = "\x01\x04"
	internalAdventureStartText = "\x01\x05"
	internalRoomName           = "\x01\x06"
	internalEndProperty        = "\x01\x07"
	internalDescription        = "\x01\x08"
	internalType               = "\x01\x09"
	internalRoomType           = "\x01\x0A"
	internalDescriptionsType   = "\x01\x0B"
	internalDescsFE            = "\x01\x0C"
	internalDescsDefault       = "\x01\x0D"
	internalDescsPickedUp      = "\x01\x0E"
	internalDescsCustom        = "\x01\x0F"
	internalSet                = "\x01\x10"
	internalSep                = "\x01\x11"
	internalRoomExits          = "\x01\x12"
	internalRoomExitDescs      = "\x01\x13"
	internalRoomItems          = "\x01\x14"
	internalProps              = "\x01\x15"
	internalItemType           = "\x01\x16"
	internalItemName           = "\x01\x17"
	internalVerbType           = "\x01\x18"
	internalVerbName           = "\x01\x19"
	internalVerbArgs           = "\x01\x1A"
	internalVerbCode           = "\x01\x1B"
	internalArgumentType       = "\x01\x1C"
	internalArgumentTypeof     = "\x01\x1D"
	internalArgumentName       = "\x01\x1E"
	internalCodeType           = "\x01\x1F"
	internalCodeKeyword        = "\x01\x80"
	internalCodeArguments      = "\x01\x81"
	internalCodeChildren       = "\x01\x82"
	internalID                 = "\x01\x83"
	internalEndFile            = "\x01\x84"
	internalAdventureType      = "\x01\x85"

	internalIDString          = "ACLANG"
	internalVersionRuntimeSep = "\x00\x00"
	internalEndRuntime        = "\x00\x01"
)

type Version struct {
	Major int
	Minor int
	Patch int
}

func (version Version) Equal(ver Version) bool {
	return version.Major == ver.Major && version.Minor == ver.Minor && version.Patch == ver.Patch
}

func (version Version) Greater(ver Version) bool {
	return ver.Major > version.Major || ver.Minor > version.Minor || ver.Patch > version.Patch
}

func (version Version) Lower(ver Version) bool {
	return ver.Major < version.Major && ver.Minor < version.Minor && ver.Patch < version.Patch
}

func (version Version) GetSemVer() string {
	return fmt.Sprintf("v%d.%d.%d", version.Major, version.Minor, version.Patch)
}

func (version Version) String() string {
	return version.GetSemVer()
}

func VersionFromSemVer(ver string) Version {
	ver = strings.TrimPrefix(ver, "v")
	var split = strings.Split(ver, ".")
	if len(split) != 3 {
		panic("invalid SemVer")
	}
	var newVersion = Version{}
	var err error
	newVersion.Major, err = strconv.Atoi(split[0])
	if err != nil {
		panic(fmt.Errorf("invalid major semver(strconv.Atoi returned %w)", err))
	}
	newVersion.Minor, err = strconv.Atoi(split[1])
	if err != nil {
		panic(fmt.Errorf("invalid minor semver(strconv.Atoi returned %w)", err))
	}
	newVersion.Patch, err = strconv.Atoi(split[2])
	if err != nil {
		panic(fmt.Errorf("invalid patch semver(strconv.Atoi returned %w)", err))
	}
	return newVersion
}

type SupportedVersion struct {
	BaseVersion Version
	ToVersion   Version
	OrHigher    bool
	OrLower     bool
}

func (version SupportedVersion) Matches(ver Version) bool {
	if version.BaseVersion.Equal(ver) || version.ToVersion.Equal(ver) {
		return true
	}
	if version.BaseVersion.Equal(version.ToVersion) && (!version.OrHigher && !version.OrLower) {
		return false
	}
	if version.BaseVersion.Greater(ver) && version.ToVersion.Lower(ver) {
		return true
	}
	if version.OrHigher && version.ToVersion.Greater(ver) {
		return true
	}
	if version.OrLower && version.BaseVersion.Lower(ver) {
		return true
	}
	return false
}

func (version SupportedVersion) String() string {
	var out = version.BaseVersion.String()
	if version.OrLower {
		out = "-" + out
	}
	if !version.BaseVersion.Equal(version.ToVersion) {
		out += "-" + version.ToVersion.String()
	}
	if version.OrHigher {
		out = out + "+"
	}
	return out
}

func compile() {
	logger.Logf(logger.LogInfo, "ACLang %v", AclangVersion)
}
