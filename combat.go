package main

import (
	"fmt"
	"math/rand"
)

type CombatState struct {
	EnemyID  string
	Enemy    EnemySnapshot
	PlayerHP int
	Turn     int
	Resolved bool
	Outcome  string
}

type EnemySnapshot struct {
	Name       string
	HP         int
	MinDamage  int
	MaxDamage  int
	WantedGain int
	FleeChance float64
}

func NewCombatState(enemyID string, enemy *Enemy) *CombatState {
	return &CombatState{
		EnemyID: enemyID,
		Enemy: EnemySnapshot{
			Name:       enemy.Name,
			HP:         enemy.HP,
			MinDamage:  enemy.MinDamage,
			MaxDamage:  enemy.MaxDamage,
			WantedGain: enemy.WantedGain,
			FleeChance: enemy.FleeChance,
		},
		Turn: 1,
	}
}

func (c *CombatState) PlayerAttack(state *GameState) string {
	if c.Resolved {
		return "Combat already resolved."
	}
	accuracy := 0.65
	if state.Player.ActiveFruit == "gale_fruit" {
		accuracy = 0.8
	}
	if rand.Float64() < accuracy {
		dmg := rand.Intn(4) + 3 + state.Player.Grit
		if state.Player.ActiveFruit == "spark_fruit" {
			dmg += 2
		}
		c.Enemy.HP -= dmg
		if c.Enemy.HP <= 0 {
			c.Resolved = true
			c.Outcome = "enemy_down"
			state.Wanted += c.Enemy.WantedGain
			return fmt.Sprintf("You strike true for %d damage. %s collapses.", dmg, c.Enemy.Name)
		}
		return fmt.Sprintf("You hit for %d damage.", dmg)
	}
	return "You miss and stumble."
}

func (c *CombatState) EnemyAttack(state *GameState) string {
	if c.Resolved {
		return ""
	}
	if rand.Float64() < c.Enemy.FleeChance {
		c.Resolved = true
		c.Outcome = "enemy_fled"
		return fmt.Sprintf("%s flees into the shadows.", c.Enemy.Name)
	}
	if rand.Float64() < 0.5 {
		dmg := rand.Intn(c.Enemy.MaxDamage-c.Enemy.MinDamage+1) + c.Enemy.MinDamage
		if state.Player.ActiveFruit == "stone_fruit" {
			dmg = max(1, dmg-2)
		}
		state.Player.HP -= dmg
		return fmt.Sprintf("%s hits you for %d damage.", c.Enemy.Name, dmg)
	}
	return fmt.Sprintf("%s swings wide.", c.Enemy.Name)
}
