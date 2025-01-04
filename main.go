package main

import (
	"errors"
	"fmt"
	"strings"
	"sync"
)

type Player struct {
	name    string
	mapID   int
	channel chan string
	game    *Game
	mu      sync.RWMutex
}

type Map struct {
	id       int
	players  map[string]*Player
	messages chan string
	mu       sync.RWMutex
}

type Game struct {
	players map[string]*Player
	maps    map[int]*Map
	mu      sync.RWMutex
}

func NewGame(mapIds []int) (*Game, error) {
	game := &Game{
		players: make(map[string]*Player),
		maps:    make(map[int]*Map),
	}

	for _, id := range mapIds {
		if id <= 0 {
			return nil, errors.New("map ID must be greater than 0")
		}
		game.maps[id] = &Map{
			id:       id,
			players:  make(map[string]*Player),
			messages: make(chan string, 100),
			mu:       sync.RWMutex{},
		}
		go game.maps[id].FanOutMessages()
	}
	return game, nil
}

func (g *Game) ConnectPlayer(name string) error {
	loweredName := strings.ToLower(name)
	
	g.mu.Lock()
	defer g.mu.Unlock()
	
	if _, exists := g.players[loweredName]; exists {
		return errors.New("player name already exists")
	}

	player := &Player{
		name:    name,
		channel: make(chan string, 100),
		mapID:   0,
		game:    g,
		mu:      sync.RWMutex{},
	}

	g.players[loweredName] = player
	return nil
}

func (g *Game) SwitchPlayerMap(name string, mapId int) error {
	g.mu.RLock()
	player, exists := g.players[strings.ToLower(name)]
	if !exists {
		g.mu.RUnlock()
		return errors.New("player not found")
	}
	targetMap, mapExists := g.maps[mapId]
	if !mapExists {
		g.mu.RUnlock()
		return errors.New("map does not exist")
	}
	g.mu.RUnlock()

	player.mu.Lock()
	if player.mapID == mapId {
		player.mu.Unlock()
		return errors.New("player is already in this map")
	}

	oldMapID := player.mapID
	player.mapID = mapId
	player.mu.Unlock()

	if oldMapID != 0 {
		oldMap := g.maps[oldMapID]
		oldMap.mu.Lock()
		delete(oldMap.players, strings.ToLower(name))
		oldMap.mu.Unlock()
	}

	targetMap.mu.Lock()
	targetMap.players[strings.ToLower(name)] = player
	targetMap.mu.Unlock()

	return nil
}

func (g *Game) GetPlayer(name string) (*Player, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	player, exists := g.players[strings.ToLower(name)]
	if !exists {
		return nil, errors.New("player not found")
	}
	return player, nil
}

func (g *Game) GetMap(mapId int) (*Map, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	m, exists := g.maps[mapId]
	if !exists {
		return nil, errors.New("map not found")
	}
	return m, nil
}

func (m *Map) FanOutMessages() {
	for msg := range m.messages {
		m.mu.RLock()
		for _, player := range m.players {
			select {
			case player.channel <- msg:
			default:
				// Channel is full, skip this message
			}
		}
		m.mu.RUnlock()
	}
}

func (p *Player) GetChannel() <-chan string {
	return p.channel
}

func (p *Player) SendMessage(msg string) error {
	p.mu.RLock()
	currentMapID := p.mapID
	p.mu.RUnlock()

	if currentMapID == 0 {
		return errors.New("player is not in any map")
	}

	p.game.mu.RLock()
	m, exists := p.game.maps[currentMapID]
	p.game.mu.RUnlock()

	if !exists {
		return errors.New("map not found")
	}

	formattedMsg := fmt.Sprintf("%s says: %s", strings.Title(strings.ToLower(p.name)), msg)
	select {
	case m.messages <- formattedMsg:
		return nil
	default:
		return errors.New("message channel is full")
	}
}

func (p *Player) GetName() string {
	return p.name
}