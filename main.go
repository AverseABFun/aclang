package main

import (
	"crypto/sha512"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/averseabfun/logger"
	"github.com/fatih/structs"
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

type Descriptions struct {
	FirstEntry   string
	Default      string
	WhenPickedUp map[*Item]string
	Custom       map[string]string
	GameObject
}

type Direction int

const (
	DirectionNorth Direction = 0
	DirectionSouth Direction = 1
	DirectionEast  Direction = 2
	DirectionWest  Direction = 3
	DirectionUp    Direction = 4
	DirectionDown  Direction = 5
)

type ID uint32

var largestID ID = math.MaxUint32

func createID() ID {
	largestID++
	if largestID == math.MaxUint32 {
		panic(fmt.Sprintf("too many game objects - attempted to create id for %dth one", math.MaxUint32))
	}
	return ID(largestID)
}

type GameObject interface {
	GetID() ID
	AssignID(ID, Adventure)
	HasSetID() bool
	GetProperties() map[string]Serializable
	GetHash() string
}

type Serializable interface {
	String() string
	GameObject
}

type Item interface {
	GetName() string
	GameObject
}

type Room interface {
	GetTitle() string
	GetDescriptions() Descriptions
	GetExits() map[Direction]ID
	GetExitDescriptions() map[Direction]Descriptions
	GetItems() []ID
	GameObject
}

type Keyword string

const (
	KeywordIf         Keyword = "\xC0\x00"
	KeywordSay        Keyword = "\xC0\x01"
	KeywordEndFail    Keyword = "\xC0\x02"
	KeywordEndSuccess Keyword = "\xC0\x03"
	KeywordDeleteItem Keyword = "\xC0\x04"
	KeywordCreateItem Keyword = "\xC0\x05"
	KeywordTeleport   Keyword = "\xC0\x06"
)

type Code interface {
	GetKeyword() Keyword
	GetArguments() []string
	GetChildren() []*Code
	GameObject
}

type ArgumentType string

const (
	ArgumentTypeItem        ArgumentType = "\xA0\x00"
	ArgumentTypeString      ArgumentType = "\xA0\x01"
	ArgumentTypePlaceholder ArgumentType = "\xA0\x02"
	ArgumentTypeRoom        ArgumentType = "\xA0\x03"
	ArgumentTypeDirection   ArgumentType = "\xA0\x04"
)

type Argument interface {
	GetType() ArgumentType
	GetName() string
	GameObject
}

type Verb interface {
	GetVerbName() string
	GetArguments() []*Argument

	GetCode() []*Code
	GameObject
}

type Adventure interface {
	GetName() string
	GetStartingText() string
	GetSupportedVersions() *SupportedVersion
	GetRooms() []Room
	GetRoomsByID() map[ID]Room
	GetStartingRoom() ID
	GetItems() []Item
	GetVerbs() []Verb
	GetAllGameObjects() []*GameObject // This should NOT include the Adventure object itself.
	GameObject
}

func GetString(obj GameObject) string {
	var out = ""
	out += internalID
	out += strconv.Itoa(int(obj.GetID()))
	out += internalEndProperty
	out += internalType
	switch v := obj.(type) {
	case Room:
		out += internalRoomType
		out += internalRoomName
		out += v.GetTitle()
		out += internalEndProperty
		out += internalDescription
		out += strconv.Itoa(int(v.GetID()))
		out += internalEndProperty
		out += internalRoomExits
		for key, val := range v.GetExits() {
			out += strconv.Itoa(int(key))
			out += internalSet
			out += strconv.Itoa(int(val))
			out += internalSep
		}
		out = out[:len(out)-len(internalSep)]
		out += internalEndProperty

		out += internalRoomExitDescs
		for key, val := range v.GetExitDescriptions() {
			out += strconv.Itoa(int(key))
			out += internalSet
			out += strconv.Itoa(int(val.GetID()))
			out += internalSep
		}
		out = out[:len(out)-len(internalSep)]
		out += internalEndProperty

		out += internalRoomItems
		for key, val := range v.GetItems() {
			out += strconv.Itoa(int(key))
			out += internalSet
			out += strconv.Itoa(int(val))
			out += internalSep
		}
		out = out[:len(out)-len(internalSep)]
		out += internalEndProperty
	case Argument:
		out += internalArgumentType
		out += internalArgumentTypeof
		out += string(v.GetType())
		out += internalEndProperty
		out += internalArgumentName
		out += v.GetName()
		out += internalEndProperty
	case Serializable:
		out = v.String()
	case Code:
		out += internalCodeType
		out += internalCodeKeyword
		out += string(v.GetKeyword())
		out += internalEndProperty
		out += internalCodeArguments
		for _, val := range v.GetArguments() {
			out += val
			out += internalSep
		}
		out = out[:len(out)-len(internalSep)]
		out += internalEndProperty
		out += internalCodeChildren
		for _, val := range v.GetChildren() {
			out += strconv.Itoa(int((*val).GetID()))
			out += internalSep
		}
		out = out[:len(out)-len(internalSep)]
		out += internalEndProperty
	case Item:
		out += internalItemType
		out += internalItemName
		out += v.GetName()
		out += internalEndProperty
	case Verb:
		out += internalVerbType
		out += internalVerbName
		out += v.GetVerbName()
		out += internalEndProperty
		out += internalVerbArgs
		for _, val := range v.GetArguments() {
			out += strconv.Itoa(int((*val).GetID()))
			out += internalSep
		}
		out = out[:len(out)-len(internalSep)]
		out += internalEndProperty
		out += internalVerbCode
		for _, val := range v.GetCode() {
			out += strconv.Itoa(int((*val).GetID()))
			out += internalSep
		}
		out = out[:len(out)-len(internalSep)]
		out += internalEndProperty
	case Descriptions:
		out += internalDescriptionsType
		out += internalDescsFE
		out += v.FirstEntry
		out += internalEndProperty
		out += internalDescsDefault
		out += v.Default
		out += internalEndProperty
		out += internalDescsPickedUp
		for key, val := range v.WhenPickedUp {
			out += strconv.Itoa(int((*key).GetID()))
			out += internalSet
			out += val
			out += internalSep
		}
		out = out[:len(out)-len(internalSep)]
		out += internalEndProperty
		out += internalDescsCustom
		for key, val := range v.Custom {
			out += key
			out += internalSet
			out += val
			out += internalSep
		}
		out = out[:len(out)-len(internalSep)]
		out += internalEndProperty
	default:
		panic(fmt.Sprintf("unknown GameObject of type %T", v))
	}
	out += internalProps
	for key, val := range obj.GetProperties() {
		out += key
		out += internalSet
		out += val.String()
		out += internalSep
	}
	out = out[:len(out)-len(internalSep)]
	out += internalEndProperty
	out += internalEndCurLookup
	return out
}

func Compile(adv Adventure) []byte {
	logger.Logf(logger.LogInfo, "ACLang %s", AclangVersion)
	if !adv.GetSupportedVersions().Matches(AclangVersion) {
		panic(fmt.Sprintf("current ACLang version %s does not match %s", AclangVersion.String(), adv.GetSupportedVersions().String()))
	}
	if !adv.HasSetID() {
		adv.AssignID(createID(), adv)
	}
	for _, obj := range adv.GetAllGameObjects() {
		if !(*obj).HasSetID() {
			(*obj).AssignID(createID(), adv)
		}
	}
	logger.Logf(logger.LogInfo, "Assigned IDs for all %d game objects", largestID)
	var out = internalIDString + AclangVersion.String() + internalVersionRuntimeSep + SupportedRuntimes.String() + internalEndRuntime
	out += internalStartingRoom
	out += strconv.Itoa(int(adv.GetStartingRoom()))
	out += internalEndCurLookup
	out += internalAdventureName
	out += adv.GetName()
	out += internalEndCurLookup
	out += internalAdventureStartText
	out += adv.GetStartingText()
	out += internalEndCurLookup

	out += internalBeginLookup
	for _, obj := range adv.GetAllGameObjects() {
		out += GetString(*obj)
	}
	out += internalEndLookup

	return ([]byte)(out)
}

type BasicGameObject struct {
	ID         ID
	setID      bool
	Properties map[string]Serializable
	AssignHook func(ID, Adventure, GameObject)
}

func (obj BasicGameObject) GetID() ID {
	return obj.ID
}

func (obj BasicGameObject) HasSetID() bool {
	return obj.setID
}

func (obj *BasicGameObject) AssignID(id ID, adv Adventure) {
	obj.ID = id
	obj.setID = true
	if obj.AssignHook != nil {
		obj.AssignHook(id, adv, GameObject(obj))
	}
}

func (obj BasicGameObject) GetHash() string {
	var data = fmt.Sprintf("%v", structs.Map(obj))
	var hasher = sha512.New()
	return string(hasher.Sum(([]byte)(data)))
}

func (obj BasicGameObject) GetProperties() map[string]Serializable {
	return obj.Properties
}

type BasicRoom struct {
	title            string
	descriptions     Descriptions
	exits            map[Direction]ID
	exitDescriptions map[Direction]Descriptions
	items            []ID
	*BasicGameObject
}

func (room BasicRoom) GetTitle() string {
	return room.title
}

func (room *BasicRoom) SetTitle(title string) {
	room.title = title
}

func (room BasicRoom) GetDescriptions() Descriptions {
	return room.descriptions
}

func (room *BasicRoom) SetDescriptions(descriptions Descriptions) {
	room.descriptions = descriptions
}

func (room BasicRoom) GetExits() map[Direction]ID {
	return room.exits
}

func (room *BasicRoom) SetExits(exits map[Direction]ID) {
	room.exits = exits
}

func (room BasicRoom) GetExitDescriptions() map[Direction]Descriptions {
	return room.exitDescriptions
}

func (room *BasicRoom) SetExitDescriptions(exitDescriptions map[Direction]Descriptions) {
	room.exitDescriptions = exitDescriptions
}

func (room BasicRoom) GetItems() []ID {
	return room.items
}

func (room *BasicRoom) SetItems(items []ID) {
	room.items = items
}

type BasicItem struct {
	name string
	*BasicGameObject
}

func (item BasicItem) GetName() string {
	return item.name
}

func (item *BasicItem) SetName(name string) {
	item.name = name
}

type BasicCode struct {
	keyword   Keyword
	arguments []string
	children  []*Code
	*BasicGameObject
}

func (code BasicCode) GetKeyword() Keyword {
	return code.keyword
}

func (code *BasicCode) SetKeyword(keyword Keyword) {
	code.keyword = keyword
}

func (code BasicCode) GetArguments() []string {
	return code.arguments
}

func (code *BasicCode) SetArguments(arguments []string) {
	code.arguments = arguments
}

func (code BasicCode) GetChildren() []*Code {
	return code.children
}

func (code *BasicCode) SetChildren(children []*Code) {
	code.children = children
}

type BasicArgument struct {
	typeof ArgumentType
	name   string
	*BasicGameObject
}

func (code BasicArgument) GetType() ArgumentType {
	return code.typeof
}

func (code *BasicArgument) SetType(typeof ArgumentType) {
	code.typeof = typeof
}

func (code BasicArgument) GetName() string {
	return code.name
}

func (code *BasicArgument) SetName(name string) {
	code.name = name
}

type BasicVerb struct {
	verbName  string
	arguments []*Argument

	code []*Code
	*BasicGameObject
}

func (verb BasicVerb) GetVerbName() string {
	return verb.verbName
}

func (verb *BasicVerb) SetVerbName(verbName string) {
	verb.verbName = verbName
}

func (verb BasicVerb) GetArguments() []*Argument {
	return verb.arguments
}

func (verb *BasicVerb) SetArguments(arguments []*Argument) {
	verb.arguments = arguments
}

func (verb BasicVerb) GetCode() []*Code {
	return verb.code
}

func (verb *BasicVerb) SetCode(code []*Code) {
	verb.code = code
}

type BasicAdventure struct {
	name              string
	startingText      string
	supportedVersions *SupportedVersion
	rooms             []Room
	roomsByID         map[ID]Room
	startingRoom      ID
	items             []Item
	verbs             []Verb
	allGameObjects    []*GameObject // This should NOT include the Adventure object itself.
	*BasicGameObject
}

func (adv BasicAdventure) GetName() string {
	return adv.name
}

func (adv *BasicAdventure) SetName(name string) {
	adv.name = name
}

func (adv BasicAdventure) GetStartingText() string {
	return adv.startingText
}

func (adv *BasicAdventure) SetStartingText(startingText string) {
	adv.startingText = startingText
}

func (adv BasicAdventure) GetSupportedVersions() *SupportedVersion {
	return adv.supportedVersions
}

func (adv *BasicAdventure) SetSupportedVersions(supportedVersions *SupportedVersion) {
	adv.supportedVersions = supportedVersions
}

func (adv BasicAdventure) GetAllGameObjects() []*GameObject {
	return adv.allGameObjects
}

func (adv BasicAdventure) GetItems() []Item {
	return adv.items
}

func (adv BasicAdventure) GetRooms() []Room {
	return adv.rooms
}

func (adv BasicAdventure) GetRoomsByID() map[ID]Room {
	return adv.roomsByID
}

func (adv BasicAdventure) GetStartingRoom() ID {
	return adv.startingRoom
}

func (adv *BasicAdventure) SetStartingRoom(room Room) {
	if !room.HasSetID() {
		room.AssignID(createID(), adv)
	}
	adv.startingRoom = room.GetID()
}

func (adv BasicAdventure) GetVerbs() []Verb {
	return adv.verbs
}

func (adv *BasicAdventure) RegisterGameObject(obj GameObject) {
	adv.allGameObjects = append(adv.allGameObjects, &obj)
	switch v := (obj).(type) {
	case Room:
		adv.rooms = append(adv.rooms, v)
		if val, ok := (obj).(*BasicRoom); ok {
			val.AssignHook = func(i ID, a Adventure, gameObj GameObject) {
				if val2, ok := a.(*BasicAdventure); ok {
					if room, ok := gameObj.(BasicRoom); ok {
						val2.roomsByID[i] = room
					}
				}
			}
		}
	case Item:
		adv.items = append(adv.items, v)
	case Verb:
		adv.verbs = append(adv.verbs, v)
	}
}

func CreateBasicAdventure() *BasicAdventure {
	return &BasicAdventure{BasicGameObject: &BasicGameObject{setID: false}}
}

func CreateBasicRoom() *BasicRoom {
	return &BasicRoom{BasicGameObject: &BasicGameObject{setID: false}}
}

func CreateBasicVerb() *BasicVerb {
	return &BasicVerb{BasicGameObject: &BasicGameObject{setID: false}}
}

func CreateBasicItem() *BasicItem {
	return &BasicItem{BasicGameObject: &BasicGameObject{setID: false}}
}

func CreateBasicArgument() *BasicArgument {
	return &BasicArgument{BasicGameObject: &BasicGameObject{setID: false}}
}

func CreateBasicCode() *BasicCode {
	return &BasicCode{BasicGameObject: &BasicGameObject{setID: false}}
}

func main() {
	var adv = CreateBasicAdventure()
	adv.SetName("test")
	adv.SetStartingText("This is just a test game lol")
	adv.SetSupportedVersions(&SupportedVersion{BaseVersion: VersionFromSemVer("v0.1.0"), ToVersion: VersionFromSemVer("v0.1.0"), OrHigher: true})
	var test_room = &BasicRoom{BasicGameObject: &BasicGameObject{setID: false}}
	test_room.SetTitle("Test Room")
	test_room.SetDescriptions(Descriptions{Default: "Test"})
	adv.RegisterGameObject(test_room)
	adv.SetStartingRoom(test_room)
	os.WriteFile("test.acl", Compile(adv), 0700)
}
