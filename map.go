package main

type WorldRoute struct {
	From   string
	To     string
	Risk   string
	Needs  string
	Locked bool
}

type WorldMap struct {
	Nodes  map[string]MapNode
	Routes []WorldRoute
}

type MapNode struct {
	ID     string
	Name   string
	X      float64
	Y      float64
	Island string
}

func BuildWorldMap() WorldMap {
	return WorldMap{
		Nodes: map[string]MapNode{
			"Harbor Isle":   {ID: "Harbor Isle", Name: "Harbor Isle", X: 60, Y: 80, Island: "Harbor Isle"},
			"Ember Isle":    {ID: "Ember Isle", Name: "Ember Isle", X: 180, Y: 40, Island: "Ember Isle"},
			"Mist Isle":     {ID: "Mist Isle", Name: "Mist Isle", X: 40, Y: 180, Island: "Mist Isle"},
			"Skyline Atoll": {ID: "Skyline Atoll", Name: "Skyline Atoll", X: 140, Y: 200, Island: "Skyline Atoll"},
			"Navy Bastion":  {ID: "Navy Bastion", Name: "Navy Bastion", X: 240, Y: 110, Island: "Navy Bastion"},
		},
		Routes: []WorldRoute{
			{From: "Harbor Isle", To: "Ember Isle", Risk: "Medium", Needs: "Chart"},
			{From: "Harbor Isle", To: "Mist Isle", Risk: "Low", Needs: ""},
			{From: "Mist Isle", To: "Skyline Atoll", Risk: "High", Needs: "Storm Lantern"},
			{From: "Harbor Isle", To: "Navy Bastion", Risk: "Severe", Needs: "Disguise"},
		},
	}
}

func PathCommands(rooms map[string]*Room, start, target string) []string {
	if start == target {
		return nil
	}
	type node struct {
		ID   string
		Prev string
		Dir  string
	}
	queue := []string{start}
	prev := map[string]node{}
	visited := map[string]bool{start: true}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		room := rooms[current]
		for dir, next := range room.Exits {
			if visited[next] {
				continue
			}
			visited[next] = true
			prev[next] = node{ID: next, Prev: current, Dir: dir}
			if next == target {
				queue = nil
				break
			}
			queue = append(queue, next)
		}
	}
	if _, ok := prev[target]; !ok {
		return nil
	}
	pathDirs := []string{}
	for current := target; current != start; {
		step := prev[current]
		pathDirs = append([]string{step.Dir}, pathDirs...)
		current = step.Prev
	}
	return pathDirs
}
