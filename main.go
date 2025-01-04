package main

import (
	"errors"
	"fmt"
	"strings"
)

type Player struct {
	name    string
	mapID   int
	channel chan string
}

type Map struct {
	id       int
	players  map[string]*Player
	messages chan string
}

type Game struct {
	players map[string]*Player
	maps    map[int]*Map
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
			messages: make(chan string, 100), // max 100 messages
		}
		go game.maps[id].FanOutMessages() // area of map
	}

	return game, nil
}

func (g *Game) ConnectPlayer(name string) error {
	if _, exists := g.players[strings.ToLower(name)]; exists {
		return errors.New("player name already exists")
	}

	player := &Player{
		name:    name,
		channel: make(chan string, 100), // max 100 messages
		mapID:   0, // no area exit
	}

	g.players[strings.ToLower(name)] = player
	return nil
}

func (g *Game) SwitchPlayerMap(name string, mapId int) error {
	player, err := g.GetPlayer(name)
	if err != nil {
		return err
	}

	if player.mapID == mapId {
		return errors.New("player is already in this map")
	}

	if _, exists := g.maps[mapId]; !exists {
		return errors.New("map does not exist")
	}

	// delete player on last area
	if player.mapID != 0 {
		delete(g.maps[player.mapID].players, strings.ToLower(name))
	}

	// add player to new area
	player.mapID = mapId
	g.maps[mapId].players[strings.ToLower(name)] = player

	return nil
}

func (g *Game) GetPlayer(name string) (*Player, error) {
	player, exists := g.players[strings.ToLower(name)]
	if !exists {
		return nil, errors.New("player not found")
	}
	return player, nil
}

func (g *Game) GetMap(mapId int) (*Map, error) {
	m, exists := g.maps[mapId]
	if !exists {
		return nil, errors.New("map not found")
	}
	return m, nil
}

func (m *Map) FanOutMessages() {
	for msg := range m.messages {
		for _, player := range m.players {
			player.channel <- msg
		}
	}
}

func (p *Player) GetChannel() <-chan string {
	return p.channel
}

func (p *Player) SendMessage(msg string) error {
	if p.mapID == 0 {
		return errors.New("player is not in any map")
	}

	formattedMsg := fmt.Sprintf("%s says: %s", strings.Title(strings.ToLower(p.name)), msg)
	p.channel <- formattedMsg
	return nil
}

func (p *Player) GetName() string {
	return p.name
}