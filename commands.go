package main

import (
	"strconv"
	"strings"
)

type CommandProcessor struct {
	Suggestions []string
}

func NewCommandProcessor() *CommandProcessor {
	return &CommandProcessor{Suggestions: []string{"look", "inventory", "talk", "use", "map", "save", "load"}}
}

func (c *CommandProcessor) Execute(state *GameState, input string) []string {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil
	}
	parts := strings.Fields(strings.ToLower(input))
	if len(parts) == 0 {
		return nil
	}
	verb := parts[0]
	if dir := normalizeDir(verb); dir != "" {
		return []string{state.Move(dir)}
	}

	switch verb {
	case "go", "move", "enter", "dock":
		if len(parts) < 2 {
			return []string{"Go where?"}
		}
		direction := normalizeDir(parts[1])
		if direction == "" {
			return []string{"That direction makes no sense."}
		}
		return []string{state.Move(direction)}
	case "look", "l":
		return []string{state.Look()}
	case "examine", "x":
		if len(parts) < 2 {
			return []string{"Examine what?"}
		}
		return []string{state.Examine(strings.Join(parts[1:], " "))}
	case "take", "get":
		if len(parts) < 2 {
			return []string{"Take what?"}
		}
		return []string{state.Take(strings.Join(parts[1:], " "))}
	case "drop":
		if len(parts) < 2 {
			return []string{"Drop what?"}
		}
		return []string{state.Drop(strings.Join(parts[1:], " "))}
	case "inventory", "i":
		return []string{inventoryText(state)}
	case "talk":
		if len(parts) < 2 {
			return []string{"Talk to whom?"}
		}
		return []string{state.Talk(strings.Join(parts[1:], " "))}
	case "bribe":
		if len(parts) < 2 {
			return []string{"Bribe whom?"}
		}
		return []string{state.Bribe(strings.Join(parts[1:], " "))}
	case "threaten":
		if len(parts) < 2 {
			return []string{"Threaten whom?"}
		}
		return []string{state.Threaten(strings.Join(parts[1:], " "))}
	case "use":
		if len(parts) < 2 {
			return []string{"Use what?"}
		}
		target := ""
		if strings.Contains(strings.ToLower(input), " on ") {
			segments := strings.SplitN(strings.ToLower(input), " on ", 2)
			if len(segments) == 2 {
				target = strings.TrimSpace(segments[1])
			}
		}
		return []string{state.Use(strings.Join(parts[1:], " "), target)}
	case "attack":
		if len(parts) < 2 {
			return []string{"Attack whom?"}
		}
		return []string{state.Attack(strings.Join(parts[1:], " "))}
	case "buy":
		if len(parts) < 2 {
			return []string{"Buy what?"}
		}
		return []string{state.Buy(strings.Join(parts[1:], " "))}
	case "sell":
		if len(parts) < 2 {
			return []string{"Sell what?"}
		}
		return []string{state.Sell(strings.Join(parts[1:], " "))}
	case "help":
		return []string{helpText()}
	case "map":
		return []string{"The map sits in the left panel. Click a room to travel."}
	case "save":
		return []string{state.Save("save1.json")}
	case "load":
		return []string{state.Load("save1.json")}
	case "quit", "exit":
		state.Flags["quit"] = true
		return []string{"You lower the sails and end your tale... for now."}
	default:
		return []string{"Unknown command. Type HELP for options."}
	}
}

func normalizeDir(dir string) string {
	switch dir {
	case "north", "n":
		return "north"
	case "south", "s":
		return "south"
	case "east", "e":
		return "east"
	case "west", "w":
		return "west"
	case "northeast", "ne":
		return "northeast"
	case "northwest", "nw":
		return "northwest"
	case "southeast", "se":
		return "southeast"
	case "southwest", "sw":
		return "southwest"
	default:
		return ""
	}
}

func inventoryText(state *GameState) string {
	if len(state.Player.Inventory) == 0 {
		return "Your pockets are empty."
	}
	lines := []string{"	Inventory:"}
	for _, itemID := range state.Player.Inventory {
		if item, ok := state.Items[itemID]; ok {
			lines = append(lines, "- "+item.Name)
		}
	}
	lines = append(lines, "Slots used: "+strconv.Itoa(state.InventorySlots())+"/"+strconv.Itoa(state.Player.MaxSlots))
	return strings.Join(lines, "\n")
}

func helpText() string {
	return strings.Join([]string{
		"Commands:",
		"Movement: GO NORTH, NORTH, N (also south/east/west)",
		"Actions: LOOK, EXAMINE <thing>, TAKE <item>, DROP <item>",
		"Social: TALK <npc>, BRIBE <npc>, THREATEN <npc>",
		"Use: USE <item> [ON <target>]",
		"Combat: ATTACK <enemy>",
		"Economy: BUY <item>, SELL <item>",
		"Utility: HELP, SAVE, LOAD, QUIT",
		"Goal: Collect three Glyph Stone fragments and escape with the treasure core.",
	}, "\n")
}
