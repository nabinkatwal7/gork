package main

import (
	"encoding/json"
	"os"
)

type SaveData struct {
	Player      Player
	RoomItems   map[string][]string
	RoomEnemies map[string][]string
	Flags       map[string]bool
	NPCState    map[string]string
	Wanted      int
	Morale      int
	Money       int
	Day         int
	TimeOfDay   int
	Discovered  map[string]bool
	Quests      map[string]*Quest
}

func (g *GameState) Save(filename string) string {
	data := SaveData{
		Player:      g.Player,
		RoomItems:   map[string][]string{},
		RoomEnemies: map[string][]string{},
		Flags:       g.Flags,
		NPCState:    g.NPCState,
		Wanted:      g.Wanted,
		Morale:      g.Morale,
		Money:       g.Money,
		Day:         g.Day,
		TimeOfDay:   g.TimeOfDay,
		Discovered:  g.Discovered,
		Quests:      g.Quests,
	}
	for id, room := range g.Rooms {
		data.RoomItems[id] = append([]string{}, room.Items...)
		data.RoomEnemies[id] = append([]string{}, room.Enemies...)
	}
	raw, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "Could not save game."
	}
	if err := os.WriteFile(filename, raw, 0644); err != nil {
		return "Could not write save file."
	}
	return "Game saved to " + filename
}

func (g *GameState) Load(filename string) string {
	raw, err := os.ReadFile(filename)
	if err != nil {
		return "Could not load save file."
	}
	var data SaveData
	if err := json.Unmarshal(raw, &data); err != nil {
		return "Save file corrupted."
	}
	fresh := NewGameState()
	*g = *fresh
	g.Player = data.Player
	g.Flags = data.Flags
	g.NPCState = data.NPCState
	g.Wanted = data.Wanted
	g.Morale = data.Morale
	g.Money = data.Money
	g.Day = data.Day
	g.TimeOfDay = data.TimeOfDay
	g.Discovered = data.Discovered
	g.Quests = data.Quests
	for id, items := range data.RoomItems {
		if room, ok := g.Rooms[id]; ok {
			room.Items = items
		}
	}
	for id, enemies := range data.RoomEnemies {
		if room, ok := g.Rooms[id]; ok {
			room.Enemies = enemies
		}
	}
	return "Game loaded."
}
