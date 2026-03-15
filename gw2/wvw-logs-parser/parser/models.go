package parser

type Fight struct {
	Name           string                 `json:"fightName"`
	Duration       int                    `json:"durationMS"`
	Players        []Player               `json:"players"`
	CombatData     []RawCombatEvent       `json:"combatData"`
	DamageTimeline map[int]*TimelineEntry `json:"damageTimeline"`
}

type Player struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Profession string `json:"profession"`
	Group      int    `json:"group"`

	DpsAll []struct {
		Dps    int `json:"dps"`
		Damage int `json:"damage"`
	} `json:"dpsAll"`

	BuffUptimes []Buff `json:"buffUptimes"`

	Defenses []struct {
		DownCount int `json:"downCount"`
		DeadCount int `json:"deadCount"`
	} `json:"defenses"`

	Rotation []RotationEntry `json:"rotation"`
}

type Buff struct {
	ID     int     `json:"id"`
	Uptime float64 `json:"uptime"`
}

type RotationEntry struct {
	ID     int     `json:"id"`
	Skills []Skill `json:"skills"`
}

type Skill struct {
	CastTime   int     `json:"castTime"`
	Duration   int     `json:"duration"`
	TimeGained int     `json:"timeGained"`
	Quickness  float64 `json:"quickness"`
}

type RawCombatEvent struct {
	Time   int `json:"time"`
	Src    int `json:"src"`
	Dst    int `json:"dst"`
	Value  int `json:"value"`
	IsBuff int `json:"isBuff"`
}

type DamageEvent struct {
	Time   int
	Source int
	Target int
	Damage int
}

type TimelineEntry struct {
	Time    int
	Players map[int]*PlayerDamage
}

type PlayerDamage struct {
	PlayerID int
	Name     string
	Damage   int
}
