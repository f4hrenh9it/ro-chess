package internal

import (
	"github.com/f4hrenh9it/ro-chess/src/server/conf"
	"github.com/f4hrenh9it/ro-chess/src/server/entity"
	"github.com/name5566/leaf/log"
	"sort"
)

func (m *Board) processCombat(updatedFigures *[]entity.Figurable, updatedPlayers *[]entity.Player, clog *[]entity.CombatEvent) *string {
	var deadPlayerName *string
	log.Debug("combat processing started")
	// higher initiative figures acts first
	// first of all applying score for figures in score zone
	sort.Sort(figuresByInitiative(m.Figures))
	deadPlayerName = m.processScoringPhase(updatedFigures, updatedPlayers)
	if deadPlayerName != nil {
		return deadPlayerName
	}
	// all skills now processed ordered by initiative
	m.processSkillPhase(updatedFigures, updatedPlayers, clog)
	// all auto attack combats now processed ordered by initiative
	m.processAttackPhase(updatedFigures, clog)
	// all moving processed ordered by initiative
	m.processMovePhase(updatedFigures)
	return nil
}

func (m *Board) processSkillPhase(updatedFigures *[]entity.Figurable, updatedPlayers *[]entity.Player, clog *[]entity.CombatEvent) {
	log.Debug("processing skill phase")
	for _, f := range m.Figures {
		if f.GetSkillSet() != nil {
			f.ApplySkills()
		}
	}
}

func (m *Board) processAttackPhase(updatedFigures *[]entity.Figurable, clog *[]entity.CombatEvent) {
	log.Debug("processing attack phase")
	for _, f := range m.Figures {
		log.Debug("figure: %s, initiative: %d", f.GetName(), f.GetInitiative())
		if f.GetAlive() == false {
			log.Debug("figure is dead, skipping")
			return
		}
		var enemyX, enemyY int
		atkX, atkY := f.GetCoords()
		player := m.Players[f.GetOwnerName()]
		if player.Side == top {
			enemyX, enemyY = atkX, atkY+1
		} else {
			enemyX, enemyY = atkX, atkY-1
		}
		log.Debug("enemy coords: %d, %d", enemyX, enemyY)
		attacker := f
		defender := m.Canvas[enemyY][enemyX].Figure
		if defender != nil {
			log.Debug("conflict!")
			defX, defY := defender.GetCoords()
			for {
				aDmg := attacker.PerformAttack()
				log.Debug("attacker dmg: %d", aDmg)
				defender.SetHP(defender.GetHP() - aDmg)
				*clog = append(*clog, entity.CombatEvent{defX, defY, -aDmg, false})
				if defender.GetHP() <= 0 {
					defender.SetAlive(false)
					m.Canvas[defY][defX].Figure = nil
					log.Debug("defender on %d, %d, figure of %s dies", atkX, atkY, defender.GetOwnerName())
					*updatedFigures = append(*updatedFigures, defender)
					break
				}
				dDmg := defender.PerformAttack()
				log.Debug("defender dmg: %d", dDmg)
				attacker.SetHP(attacker.GetHP() - dDmg)
				*clog = append(*clog, entity.CombatEvent{atkX, atkY, -dDmg, false})
				if attacker.GetHP() <= 0 {
					attacker.SetAlive(false)
					m.Canvas[atkY][atkX].Figure = nil
					log.Debug("attacker on %d, %d, figure of %s dies", atkX, atkY, attacker.GetOwnerName())
					*updatedFigures = append(*updatedFigures, attacker)
					break
				}
			}
		}
	}
	m.removeDeadFigures(&m.Figures)
}

func (m *Board) processMovePhase(updatedFigures *[]entity.Figurable) {
	log.Debug("processing move phase")
	for _, newFigure := range m.Figures {
		// all active figures moves 1 cell up or down, depends on player side
		if newFigure.GetActive() == true && newFigure.GetAlive() == true {
			prevTurnX, prevTurnY := newFigure.GetCoords()
			// set prev coords so we can update all using one figure data
			newFigure.SetPrevCoords(prevTurnX, prevTurnY)
			var newX, newY int
			player := m.Players[newFigure.GetOwnerName()]
			if player.Side == top {
				newX, newY = prevTurnX, prevTurnY+1
			} else {
				newX, newY = prevTurnX, prevTurnY-1
			}
			prevCell := m.Canvas[prevTurnY][prevTurnX]

			targetCell := m.Canvas[newY][newX]
			if targetCell.Figure != nil {
				log.Debug("figure target cell is not empty, skipping moving")
			}

			// all scoring and battle mechanics happens before this moment
			targetCell.Figure = newFigure
			prevCell.Figure = nil
			newFigure.SetCoords(newX, newY)
			log.Debug("moved %s from %d, %d to %d, %d", newFigure.GetName(), prevTurnX, prevTurnY, newX, newY)
			m.VisualizeAll()

			*updatedFigures = append(*updatedFigures, newFigure)
		}
	}
}

func (m *Board) processScoringPhase(figures *[]entity.Figurable, players *[]entity.Player) *string {
	// if our figure comes to opponent dmg zone, deal dmg equal to attack
	log.Debug("processing score phase")
	for _, f := range m.Figures {
		x, y := f.GetCoords()
		if y == conf.Server.ZoneScoreYTop || y == conf.Server.ZoneScoreYBottom {
			player := m.Players[f.GetOwnerName()]
			// think about damage model for players, is min-max from PerformAttack() is ok?
			magicNumber := 3
			log.Debug("%s takes damage from figure is score zone!", player.Opponent.Name)
			log.Debug("figure: %s", f.GetName())
			log.Debug("figure damage: %d", magicNumber)
			f.SetAlive(false)
			m.Canvas[y][x] = nil
			player.Opponent.HP = player.Opponent.HP - magicNumber
			*players = append(*players, entity.Player{
				Name: player.Opponent.Name,
				HP:   player.Opponent.HP,
				MP:   player.Opponent.MP,
			},
			)
			f.SetAlive(false)
			*figures = append(*figures, f)
			// if it's squad, do the math and replace all squad units in one time
		}
	}
	m.removeDeadFigures(&m.Figures)
	for _, p := range m.Players {
		if p.HP <= 0 {
			return &p.Name
		}
	}
	return nil
}

func (m *Board) removeDeadFigures(bfs *[]entity.Figurable) {
	for i, bf := range *bfs {
		if bf.GetAlive() == false {
			*bfs = append((*bfs)[:i], (*bfs)[i+1:]...)
		}
	}
}
