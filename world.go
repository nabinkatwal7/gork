package main

import (
	"fmt"
	"math/rand"
	"strings"
)

type Player struct {
	Location    string
	Inventory   []string
	Equipped    map[string]string
	MaxSlots    int
	HP          int
	MaxHP       int
	Grit        int
	Charm       int
	Wits        int
	ActiveFruit string
}

type GameState struct {
	Rooms      map[string]*Room
	Items      map[string]*Item
	NPCs       map[string]*NPC
	Enemies    map[string]*Enemy
	Quests     map[string]*Quest
	Islands    map[string]*Island
	Player     Player
	Flags      map[string]bool
	NPCState   map[string]string
	Wanted     int
	Morale     int
	Money      int
	Day        int
	TimeOfDay  int
	Log        []LogEntry
	Combat     *CombatState
	Discovered map[string]bool
}

type LogEntry struct {
	Time string
	Text string
	Kind string
}

func NewGameState() *GameState {
	rooms, items, npcs, enemies, quests, islands := WorldData()
	state := &GameState{
		Rooms:      rooms,
		Items:      items,
		NPCs:       npcs,
		Enemies:    enemies,
		Quests:     quests,
		Islands:    islands,
		Player:     Player{Location: "ship_deck", Inventory: []string{}, Equipped: map[string]string{"weapon": "", "charm": "", "tool": ""}, MaxSlots: 12, HP: 24, MaxHP: 24, Grit: 2, Charm: 2, Wits: 2},
		Flags:      map[string]bool{},
		NPCState:   map[string]string{},
		Wanted:     0,
		Morale:     0,
		Money:      80,
		Day:        1,
		TimeOfDay:  9,
		Log:        []LogEntry{},
		Discovered: map[string]bool{},
	}
	state.MarkDiscovered("ship_deck")
	state.MarkDiscovered("ship_cabin")
	state.AddLog("You are a rookie captain chasing legendary treasure across the Wild Current.", "story")
	state.AddLog("Try LOOK, INVENTORY, and GO NORTH to begin.", "hint")
	for id, npc := range state.NPCs {
		state.NPCState[id] = npc.Disposition
	}
	return state
}

func (g *GameState) AddLog(text string, kind string) {
	g.Log = append([]LogEntry{{Time: g.TimeStamp(), Text: text, Kind: kind}}, g.Log...)
}

func (g *GameState) TimeStamp() string {
	return fmt.Sprintf("Day %d %02d:00", g.Day, g.TimeOfDay)
}

func (g *GameState) AdvanceTime() {
	g.TimeOfDay++
	if g.TimeOfDay >= 24 {
		g.Day++
		g.TimeOfDay = 0
	}
}

func (g *GameState) Room() *Room {
	return g.Rooms[g.Player.Location]
}

func (g *GameState) MarkDiscovered(id string) {
	if g.Discovered[id] {
		return
	}
	g.Discovered[id] = true
	if room, ok := g.Rooms[id]; ok {
		room.Discovered = true
	}
}

func (g *GameState) InventorySlots() int {
	count := 0
	for _, itemID := range g.Player.Inventory {
		count += g.Items[itemID].Slots
	}
	return count
}

func (g *GameState) HasItem(id string) bool {
	for _, itemID := range g.Player.Inventory {
		if itemID == id {
			return true
		}
	}
	return false
}

func (g *GameState) FindItem(name string, list []string) string {
	name = strings.ToLower(name)
	for _, itemID := range list {
		item := g.Items[itemID]
		if strings.ToLower(item.Name) == name || itemID == name {
			return itemID
		}
	}
	return ""
}

func (g *GameState) FindNPC(name string, list []string) string {
	name = strings.ToLower(name)
	for _, npcID := range list {
		npc := g.NPCs[npcID]
		if strings.ToLower(npc.Name) == name || npcID == name {
			return npcID
		}
	}
	return ""
}

func (g *GameState) FindEnemy(name string, list []string) string {
	name = strings.ToLower(name)
	for _, enemyID := range list {
		enemy := g.Enemies[enemyID]
		if strings.ToLower(enemy.Name) == name || enemyID == name {
			return enemyID
		}
	}
	return ""
}

func (g *GameState) Move(direction string) string {
	room := g.Room()
	if room == nil {
		return "You are lost in the Wild Current."
	}
	dest, ok := room.Exits[direction]
	if !ok {
		return "You can't go that way."
	}
	if err := g.CanEnter(dest); err != "" {
		return err
	}
	g.Player.Location = dest
	g.MarkDiscovered(dest)
	g.AdvanceTime()
	g.MaybePatrol()
	return g.Look()
}

func (g *GameState) MaybePatrol() {
	if g.Wanted < 3 {
		return
	}
	room := g.Room()
	if room == nil || len(room.Enemies) > 0 || room.Island == "Ship" {
		return
	}
	if rand.Float64() < 0.3 {
		room.Enemies = append(room.Enemies, "navy_patrol")
		g.AddLog("A Bluecoat patrol storms in, nets ready.", "event")
	}
}

func (g *GameState) CanEnter(dest string) string {
	if dest == "navy_outpost" && !g.Flags["bribed"] {
		return "The Bluecoat officer blocks the way. A donation might help."
	}
	if dest == "ruins_hall" && !g.Flags["ruinUnlocked"] {
		return "The stone gate is locked."
	}
	if dest == "ruins_core" && !g.Flags["innerUnlocked"] {
		return "A sealed door bars the way. The sea must hear your call."
	}
	if dest == "reef_shallows" && g.Player.ActiveFruit != "" {
		g.Flags["drowned"] = true
		return "The cursed power drags you under the waves. The sea refuses you."
	}
	if dest == "sky_shrine" && g.Player.ActiveFruit == "stone_fruit" {
		return "The stone curse makes the storm lift impossible. You're too heavy."
	}
	if g.Wanted >= 5 && dest == "navy_outpost" {
		return "Bluecoat Navy seals the outpost. You're turned away."
	}
	return ""
}

func (g *GameState) Look() string {
	room := g.Room()
	if room == nil {
		return "You see nothing but mist."
	}
	lines := []string{fmt.Sprintf("%s - %s", room.Name, room.Island), room.Desc}
	if len(room.Items) > 0 {
		lines = append(lines, "You see: "+g.ListItemNames(room.Items))
	}
	if len(room.NPCs) > 0 {
		lines = append(lines, "People here: "+g.ListNPCNames(room.NPCs))
	}
	if len(room.Enemies) > 0 {
		lines = append(lines, "Threats: "+g.ListEnemyNames(room.Enemies))
	}
	lines = append(lines, "Exits: "+strings.Join(exitKeys(room.Exits), ", "))
	return strings.Join(lines, "\n")
}

func (g *GameState) ListItemNames(ids []string) string {
	names := make([]string, 0, len(ids))
	for _, id := range ids {
		if item, ok := g.Items[id]; ok {
			names = append(names, item.Name)
		}
	}
	return strings.Join(names, ", ")
}

func (g *GameState) ListNPCNames(ids []string) string {
	names := make([]string, 0, len(ids))
	for _, id := range ids {
		if npc, ok := g.NPCs[id]; ok {
			names = append(names, npc.Name)
		}
	}
	return strings.Join(names, ", ")
}

func (g *GameState) ListEnemyNames(ids []string) string {
	names := make([]string, 0, len(ids))
	for _, id := range ids {
		if enemy, ok := g.Enemies[id]; ok {
			names = append(names, enemy.Name)
		}
	}
	return strings.Join(names, ", ")
}

func exitKeys(exits map[string]string) []string {
	keys := make([]string, 0, len(exits))
	for key := range exits {
		keys = append(keys, key)
	}
	return keys
}

func (g *GameState) Examine(name string) string {
	if itemID := g.FindItem(name, g.Player.Inventory); itemID != "" {
		return g.Items[itemID].Desc
	}
	room := g.Room()
	if itemID := g.FindItem(name, room.Items); itemID != "" {
		return g.Items[itemID].Desc
	}
	if npcID := g.FindNPC(name, room.NPCs); npcID != "" {
		return g.NPCs[npcID].Desc
	}
	if enemyID := g.FindEnemy(name, room.Enemies); enemyID != "" {
		return g.Enemies[enemyID].Desc
	}
	return "You find nothing like that to examine."
}

func (g *GameState) Take(name string) string {
	room := g.Room()
	itemID := g.FindItem(name, room.Items)
	if itemID == "" {
		return "You don't see that here."
	}
	item := g.Items[itemID]
	if g.InventorySlots()+item.Slots > g.Player.MaxSlots {
		return "You're carrying too much already."
	}
	room.Items = removeID(room.Items, itemID)
	g.Player.Inventory = append(g.Player.Inventory, itemID)
	if item.Contraband {
		g.Wanted++
	}
	return fmt.Sprintf("You take the %s.", item.Name)
}

func (g *GameState) Drop(name string) string {
	itemID := g.FindItem(name, g.Player.Inventory)
	if itemID == "" {
		return "You don't have that."
	}
	g.Player.Inventory = removeID(g.Player.Inventory, itemID)
	room := g.Room()
	room.Items = append(room.Items, itemID)
	return fmt.Sprintf("You drop the %s.", g.Items[itemID].Name)
}

func (g *GameState) Talk(name string) string {
	room := g.Room()
	npcID := g.FindNPC(name, room.NPCs)
	if npcID == "" {
		return "No one like that is here."
	}
	if g.NPCState[npcID] == "hostile" {
		return "They glare and refuse to speak."
	}
	npc := g.NPCs[npcID]
	response := npc.Talk
	if g.Wanted >= 4 && npc.Disposition == "hostile" {
		response = "The Bluecoat glowers. 'Hands where I can see them.'"
	}
	if g.NPCState[npcID] == "neutral" {
		g.NPCState[npcID] = "met"
	}
	return response
}

func (g *GameState) Bribe(name string) string {
	room := g.Room()
	npcID := g.FindNPC(name, room.NPCs)
	if npcID == "" {
		return "There's no one here to bribe."
	}
	if g.Money < 25 {
		return "You don't have enough coin to bribe convincingly."
	}
	g.Money -= 25
	g.Wanted = max(0, g.Wanted-1)
	g.Flags["bribed"] = true
	g.NPCState[npcID] = "friendly"
	return "The bribe slips into a pocket. The way is suddenly less guarded."
}

func (g *GameState) Threaten(name string) string {
	room := g.Room()
	npcID := g.FindNPC(name, room.NPCs)
	if npcID == "" {
		return "No one here looks threatened."
	}
	check := g.SkillCheck("grit")
	if check {
		g.Wanted++
		g.NPCState[npcID] = "hostile"
		return "Your threat lands. People scatter and the wanted posters multiply."
	}
	g.Wanted++
	g.NPCState[npcID] = "hostile"
	return "Your threat falls flat. Someone laughs."
}

func (g *GameState) Use(itemName string, target string) string {
	itemID := g.FindItem(itemName, g.Player.Inventory)
	if itemID == "" {
		return "You don't have that to use."
	}
	item := g.Items[itemID]
	if item.Fruit {
		if g.Player.ActiveFruit != "" {
			return "Only one cursed fruit at a time. The sea insists."
		}
		g.Player.ActiveFruit = itemID
		g.Player.Inventory = removeID(g.Player.Inventory, itemID)
		g.Morale++
		return "Power surges through you. The sea now resents you."
	}
	switch itemID {
	case "cipher_lens":
		if g.Player.Location != "mist_library" {
			return "The lens needs a quiet library to read the glyphs."
		}
		if g.HasItem("glyph_frag_1") && g.HasItem("glyph_frag_2") && g.HasItem("glyph_frag_3") {
			g.Flags["coordsDecoded"] = true
			g.Player.Inventory = append(g.Player.Inventory, "treasure_core")
			return "The lens reveals the Treasure Coordinate Core within the fragments."
		}
		return "The lens reveals hints, but you need all fragments."
	case "stone_key":
		if g.Player.Location == "ruins_gate" {
			g.Flags["ruinUnlocked"] = true
			return "The stone key turns. The gate groans open."
		}
	case "storm_lantern":
		if g.Player.Location == "ruins_hall" {
			g.Flags["innerUnlocked"] = true
			return "The lantern's glow wakes hidden runes. The inner door opens."
		}
	case "gadget_gull":
		g.Morale++
		return "The gull chirps. Your crew laughs. Morale rises."
	case "rum":
		if target == "broker" {
			g.Player.Inventory = removeID(g.Player.Inventory, itemID)
			g.Player.Inventory = append(g.Player.Inventory, "stone_key")
			if quest, ok := g.Quests["broker"]; ok {
				quest.Done = true
				quest.Outcome = "The broker traded a stone key."
			}
			return "The broker trades the rum for a stone key."
		}
		g.Morale++
		return "You take a sip. Courage bubbles up."
	case "medkit":
		if target == "dockhand" {
			g.Player.Inventory = removeID(g.Player.Inventory, itemID)
			g.Player.Inventory = append(g.Player.Inventory, "sun_coin")
			if quest, ok := g.Quests["dockhand"]; ok {
				quest.Done = true
				quest.Outcome = "The dockhand repaid your kindness."
			}
			return "You patch the dockhand. They slip you a sun coin."
		}
		g.Player.HP = min(g.Player.MaxHP, g.Player.HP+6)
		return "You patch yourself up."
	case "sun_coin":
		if target == "shrine" || g.Player.Location == "sky_shrine" {
			g.Morale += 2
			g.Flags["shrineBlessing"] = true
			if quest, ok := g.Quests["priest"]; ok {
				quest.Done = true
				quest.Outcome = "The shrine accepted your offering."
			}
			return "The shrine hums. The storm calms for now."
		}
	case "bribe":
		if target == "bluecoat officer" || target == "officer" {
			g.Flags["bribed"] = true
			g.Player.Inventory = removeID(g.Player.Inventory, itemID)
			return "The officer pockets the coins and steps aside."
		}
	case "spice":
		if target == "gadgeteer" {
			g.Player.Inventory = removeID(g.Player.Inventory, itemID)
			g.Player.Inventory = append(g.Player.Inventory, "cipher_lens")
			if quest, ok := g.Quests["gadgeteer"]; ok {
				quest.Done = true
				quest.Outcome = "Spice traded for a cipher lens."
			}
			return "The gadgeteer trades a cipher lens for the spice."
		}
	case "repair_kit":
		if target == "shipwright" {
			g.Player.Inventory = removeID(g.Player.Inventory, itemID)
			g.Player.Inventory = append(g.Player.Inventory, "dock_pass")
			g.Morale++
			if quest, ok := g.Quests["shipwright"]; ok {
				quest.Done = true
				quest.Outcome = "The shipwright granted you a dock pass."
			}
			return "The shipwright hands you a dock pass."
		}
	case "treasure_core":
		if g.Player.Location == "ship_deck" {
			g.Flags["treasureEscaped"] = true
			return "You set the coordinates and cut the sails."
		}
	}
	return "Nothing happens."
}

func (g *GameState) Attack(name string) string {
	room := g.Room()
	enemyID := g.FindEnemy(name, room.Enemies)
	if enemyID == "" {
		return "No enemy by that name is here."
	}
	g.Combat = NewCombatState(enemyID, g.Enemies[enemyID])
	return fmt.Sprintf("Combat begins with %s!", g.Enemies[enemyID].Name)
}

func (g *GameState) Buy(itemName string) string {
	room := g.Room()
	if !contains(room.Tags, "shop") {
		return "There's nothing for sale here."
	}
	itemID := g.FindItem(itemName, room.Items)
	if itemID == "" {
		return "That item isn't for sale here."
	}
	item := g.Items[itemID]
	price := g.Price(item.Value)
	if g.Money < price {
		return "You can't afford that."
	}
	if g.InventorySlots()+item.Slots > g.Player.MaxSlots {
		return "You're carrying too much already."
	}
	g.Money -= price
	room.Items = removeID(room.Items, itemID)
	g.Player.Inventory = append(g.Player.Inventory, itemID)
	return fmt.Sprintf("You buy %s for %d coins.", item.Name, price)
}

func (g *GameState) Sell(itemName string) string {
	room := g.Room()
	if !contains(room.Tags, "shop") {
		return "No one is buying here."
	}
	itemID := g.FindItem(itemName, g.Player.Inventory)
	if itemID == "" {
		return "You don't have that to sell."
	}
	item := g.Items[itemID]
	sale := g.Price(item.Value / 2)
	g.Money += sale
	g.Player.Inventory = removeID(g.Player.Inventory, itemID)
	room.Items = append(room.Items, itemID)
	return fmt.Sprintf("You sell %s for %d coins.", item.Name, sale)
}

func (g *GameState) Price(base int) int {
	mod := 1.0 + float64(g.Wanted)*0.05
	if g.Morale >= 3 {
		mod -= 0.1
	}
	return int(float64(base) * mod)
}

func (g *GameState) SkillCheck(stat string) bool {
	roll := rand.Intn(20) + 1
	bonus := g.Morale / 2
	statBonus := 0
	switch stat {
	case "grit":
		statBonus = g.Player.Grit
	case "charm":
		statBonus = g.Player.Charm
	case "wits":
		statBonus = g.Player.Wits
	}
	return roll+statBonus+bonus >= 12
}

func (g *GameState) ResolveQuests() {
	if g.HasItem("glyph_frag_1") && g.HasItem("glyph_frag_2") && g.HasItem("glyph_frag_3") {
		if quest, ok := g.Quests["main"]; ok {
			quest.Done = true
			quest.Outcome = "Fragments secured. Decode them with a cipher lens."
		}
	}
	if g.Flags["ruinUnlocked"] {
		if quest, ok := g.Quests["broker"]; ok {
			quest.Done = true
			quest.Outcome = "Key delivered."
		}
	}
}

func (g *GameState) EndingsCheck() (bool, string) {
	if g.Player.HP <= 0 {
		return true, "You slump to the ground. The Bluecoat Navy captures you."
	}
	if g.Flags["drowned"] {
		return true, "The sea claims you for daring its curse."
	}
	if g.Wanted >= 7 {
		return true, "Bluecoat Navy corners you. Chains clamp shut."
	}
	if g.Flags["treasureEscaped"] {
		if g.Player.ActiveFruit != "" {
			return true, "You escape with the treasure, but the curse twists your fate."
		}
		return true, "You vanish into the Wild Current with the treasure."
	}
	if g.Flags["treasureLost"] {
		return true, "The rival pirate steals the treasure core. Your legend ends in a whimper."
	}
	return false, ""
}

func removeID(list []string, id string) []string {
	result := make([]string, 0, len(list))
	for _, entry := range list {
		if entry != id {
			result = append(result, entry)
		}
	}
	return result
}

func contains(list []string, target string) bool {
	for _, entry := range list {
		if entry == target {
			return true
		}
	}
	return false
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
