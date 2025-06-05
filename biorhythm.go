package main

import (
	"math"
	"math/rand"
)

// ActivityType represents different types of activities entities can perform
type ActivityType int

const (
	ActivitySleep ActivityType = iota
	ActivityEat
	ActivityDrink
	ActivityPlay
	ActivityExplore
	ActivityScavenge
	ActivityRest
	ActivitySocialize
)

// ActivityState tracks when an entity last performed an activity and their need level
type ActivityState struct {
	LastPerformed int     `json:"last_performed"` // World tick when last performed
	NeedLevel     float64 `json:"need_level"`     // Current need level (0.0 to 1.0)
	IsActive      bool    `json:"is_active"`      // Whether currently performing this activity
	Duration      int     `json:"duration"`       // How long they've been doing this activity
}

// BioRhythm manages an entity's biological rhythms and activity needs
type BioRhythm struct {
	EntityID          int                             `json:"entity_id"`
	Activities        map[ActivityType]*ActivityState `json:"activities"`
	SleepCycles       int                             `json:"sleep_cycles"`         // Number of complete sleep cycles
	LastMealTick      int                             `json:"last_meal_tick"`       // When they last ate
	LastDrinkTick     int                             `json:"last_drink_tick"`      // When they last drank
	EnergyAtLastSleep float64                         `json:"energy_at_last_sleep"` // Energy level when they last slept
	ActivitySchedule  map[TimeOfDay][]ActivityType    `json:"activity_schedule"`    // Preferred activities by time of day
	CircadianClock    float64                         `json:"circadian_clock"`      // Internal biological clock (0.0 to 1.0)
}

// NewBioRhythm creates a new biorhythm system for an entity
func NewBioRhythm(entityID int, entity *Entity) *BioRhythm {
	br := &BioRhythm{
		EntityID:          entityID,
		Activities:        make(map[ActivityType]*ActivityState),
		ActivitySchedule:  make(map[TimeOfDay][]ActivityType),
		CircadianClock:    rand.Float64(), // Random starting point in circadian cycle
		EnergyAtLastSleep: 100.0,
	}

	// Initialize activity states
	activities := []ActivityType{
		ActivitySleep, ActivityEat, ActivityDrink, ActivityPlay,
		ActivityExplore, ActivityScavenge, ActivityRest, ActivitySocialize,
	}

	for _, activity := range activities {
		br.Activities[activity] = &ActivityState{
			LastPerformed: 0,
			NeedLevel:     rand.Float64() * 0.5, // Start with some random need
			IsActive:      false,
			Duration:      0,
		}
	}

	// Set up activity schedule based on circadian preference
	br.setupActivitySchedule(entity)

	return br
}

// setupActivitySchedule creates a schedule of preferred activities based on entity traits
func (br *BioRhythm) setupActivitySchedule(entity *Entity) {
	circadianPref := entity.GetTrait("circadian_preference")
	intelligence := entity.GetTrait("intelligence")
	cooperation := entity.GetTrait("cooperation")

	// Clear existing schedule
	for timeOfDay := Dawn; timeOfDay <= LateNight; timeOfDay++ {
		br.ActivitySchedule[timeOfDay] = []ActivityType{}
	}

	if circadianPref < -0.3 { // Nocturnal
		// Active at night
		br.ActivitySchedule[Night] = []ActivityType{ActivityExplore, ActivityScavenge, ActivityEat}
		br.ActivitySchedule[Midnight] = []ActivityType{ActivityExplore, ActivityScavenge, ActivitySocialize}
		br.ActivitySchedule[LateNight] = []ActivityType{ActivityDrink, ActivityRest}
		// Sleep during day
		br.ActivitySchedule[Morning] = []ActivityType{ActivitySleep, ActivityRest}
		br.ActivitySchedule[Midday] = []ActivityType{ActivitySleep}
		br.ActivitySchedule[Afternoon] = []ActivityType{ActivitySleep, ActivityRest}
		// Transition periods
		br.ActivitySchedule[Dawn] = []ActivityType{ActivityDrink, ActivityRest}
		br.ActivitySchedule[Evening] = []ActivityType{ActivityEat, ActivitySocialize}
	} else if circadianPref > 0.3 { // Diurnal
		// Active during day
		br.ActivitySchedule[Dawn] = []ActivityType{ActivityDrink, ActivityEat}
		br.ActivitySchedule[Morning] = []ActivityType{ActivityExplore, ActivityEat, ActivityScavenge}
		br.ActivitySchedule[Midday] = []ActivityType{ActivityEat, ActivitySocialize, ActivityPlay}
		br.ActivitySchedule[Afternoon] = []ActivityType{ActivityExplore, ActivityScavenge}
		br.ActivitySchedule[Evening] = []ActivityType{ActivityDrink, ActivitySocialize}
		// Sleep at night
		br.ActivitySchedule[Night] = []ActivityType{ActivitySleep, ActivityRest}
		br.ActivitySchedule[Midnight] = []ActivityType{ActivitySleep}
		br.ActivitySchedule[LateNight] = []ActivityType{ActivitySleep, ActivityRest}
	} else { // Crepuscular or flexible
		// Active during dawn and dusk
		br.ActivitySchedule[Dawn] = []ActivityType{ActivityEat, ActivityDrink, ActivityExplore}
		br.ActivitySchedule[Morning] = []ActivityType{ActivityRest, ActivityPlay}
		br.ActivitySchedule[Midday] = []ActivityType{ActivitySleep, ActivityRest}
		br.ActivitySchedule[Afternoon] = []ActivityType{ActivityRest, ActivityPlay}
		br.ActivitySchedule[Evening] = []ActivityType{ActivityEat, ActivityDrink, ActivityExplore}
		br.ActivitySchedule[Night] = []ActivityType{ActivityScavenge, ActivitySocialize}
		br.ActivitySchedule[Midnight] = []ActivityType{ActivitySleep, ActivityRest}
		br.ActivitySchedule[LateNight] = []ActivityType{ActivitySleep, ActivityRest}
	}

	// Add social activities for cooperative entities
	if cooperation > 0.5 {
		for timeOfDay := range br.ActivitySchedule {
			if len(br.ActivitySchedule[timeOfDay]) > 0 {
				// Add socializing to periods when already active
				br.ActivitySchedule[timeOfDay] = append(br.ActivitySchedule[timeOfDay], ActivitySocialize)
			}
		}
	}

	// Add play activities for intelligent entities
	if intelligence > 0.4 {
		for timeOfDay := range br.ActivitySchedule {
			if timeOfDay != Night && timeOfDay != Midnight && timeOfDay != LateNight {
				if len(br.ActivitySchedule[timeOfDay]) > 0 {
					br.ActivitySchedule[timeOfDay] = append(br.ActivitySchedule[timeOfDay], ActivityPlay)
				}
			}
		}
	}
}

// Update updates the biorhythm system and calculates current needs
func (br *BioRhythm) Update(tick int, entity *Entity, timeState TimeState) {
	// Update circadian clock
	br.updateCircadianClock(timeState)

	// Update activity needs
	br.updateActivityNeeds(tick, entity, timeState)

	// Update current activity durations
	br.updateActivityDurations()

	// Determine current activity based on needs and schedule
	br.determineCurrentActivity(tick, entity, timeState)
}

// updateCircadianClock updates the internal biological clock
func (br *BioRhythm) updateCircadianClock(timeState TimeState) {
	// Sync with day/night cycle gradually
	targetClock := float64(timeState.TimeOfDay) / 8.0 // 8 time periods in a day

	// Gradually sync internal clock to external time
	syncRate := 0.01 // How quickly internal clock syncs to external
	br.CircadianClock += (targetClock - br.CircadianClock) * syncRate

	// Keep clock in valid range
	if br.CircadianClock < 0 {
		br.CircadianClock += 1.0
	}
	if br.CircadianClock >= 1.0 {
		br.CircadianClock -= 1.0
	}
}

// updateActivityNeeds increases need levels over time based on entity traits
func (br *BioRhythm) updateActivityNeeds(tick int, entity *Entity, timeState TimeState) {
	// Get trait-based need rates
	sleepNeed := math.Max(0.1, entity.GetTrait("sleep_need"))
	hungerNeed := math.Max(0.1, entity.GetTrait("hunger_need"))
	thirstNeed := math.Max(0.1, entity.GetTrait("thirst_need"))
	playDrive := entity.GetTrait("play_drive")
	explorationDrive := entity.GetTrait("exploration_drive")
	scavengingBehavior := entity.GetTrait("scavenging_behavior")

	// Increase needs over time
	// Sleep need increases over time, faster if low energy
	sleepRate := sleepNeed * 0.002 // Base rate
	if entity.Energy < 50 {
		sleepRate *= 2.0 // Sleep more when tired
	}
	br.Activities[ActivitySleep].NeedLevel += sleepRate

	// Hunger increases over time and with activity
	hungerRate := hungerNeed * 0.003
	if entity.Energy < 40 {
		hungerRate *= 1.5 // Get hungrier when low energy
	}
	br.Activities[ActivityEat].NeedLevel += hungerRate

	// Thirst increases over time, faster in hot weather
	thirstRate := thirstNeed * 0.0025
	if timeState.Temperature > 0.7 {
		thirstRate *= 1.3 // Thirstier when hot
	}
	br.Activities[ActivityDrink].NeedLevel += thirstRate

	// Play drive increases if intelligent and well-fed
	if playDrive > 0 && entity.Energy > 60 {
		br.Activities[ActivityPlay].NeedLevel += math.Max(0, playDrive*0.001)
	}

	// Exploration drive increases over time if curious
	if explorationDrive > 0 {
		br.Activities[ActivityExplore].NeedLevel += math.Max(0, explorationDrive*0.0015)
	}

	// Scavenging behavior increases if low on food or naturally inclined
	if scavengingBehavior > 0 {
		scavengeRate := scavengingBehavior * 0.001
		if entity.Energy < 50 {
			scavengeRate *= 2.0 // Scavenge more when hungry
		}
		br.Activities[ActivityScavenge].NeedLevel += scavengeRate
	}

	// Rest need increases with fatigue and age
	restRate := 0.0005
	if entity.Age > 1000 {
		restRate *= 1.5 // Older entities need more rest
	}
	br.Activities[ActivityRest].NeedLevel += restRate

	// Social need increases for cooperative entities
	cooperation := entity.GetTrait("cooperation")
	if cooperation > 0 {
		br.Activities[ActivitySocialize].NeedLevel += cooperation * 0.0008
	}

	// Cap all needs at maximum
	for _, activity := range br.Activities {
		if activity.NeedLevel > 1.0 {
			activity.NeedLevel = 1.0
		}
	}
}

// updateActivityDurations updates how long the entity has been performing activities
func (br *BioRhythm) updateActivityDurations() {
	for _, activity := range br.Activities {
		if activity.IsActive {
			activity.Duration++
		} else {
			activity.Duration = 0
		}
	}
}

// determineCurrentActivity determines what the entity should be doing right now
func (br *BioRhythm) determineCurrentActivity(tick int, entity *Entity, timeState TimeState) {
	// First, stop all current activities
	for _, activity := range br.Activities {
		activity.IsActive = false
	}

	// Get scheduled activities for current time
	scheduledActivities := br.ActivitySchedule[timeState.TimeOfDay]

	// Find the activity with highest need among scheduled activities
	var bestActivity ActivityType
	highestNeed := 0.0

	// Check scheduled activities first
	for _, activity := range scheduledActivities {
		need := br.Activities[activity].NeedLevel
		if need > highestNeed {
			highestNeed = need
			bestActivity = activity
		}
	}

	// If no scheduled activity has high need, check critical needs
	if highestNeed < 0.7 {
		// Critical sleep need (override schedule)
		if br.Activities[ActivitySleep].NeedLevel > 0.8 {
			bestActivity = ActivitySleep
			highestNeed = br.Activities[ActivitySleep].NeedLevel
		}
		// Critical hunger
		if br.Activities[ActivityEat].NeedLevel > 0.9 {
			bestActivity = ActivityEat
			highestNeed = br.Activities[ActivityEat].NeedLevel
		}
		// Critical thirst
		if br.Activities[ActivityDrink].NeedLevel > 0.9 {
			bestActivity = ActivityDrink
			highestNeed = br.Activities[ActivityDrink].NeedLevel
		}
	}

	// If still no strong need, default to rest
	if highestNeed < 0.3 {
		bestActivity = ActivityRest
	}

	// Start the selected activity
	if activity := br.Activities[bestActivity]; activity != nil {
		activity.IsActive = true
		activity.Duration = 1

		// Perform activity effects
		br.performActivity(bestActivity, tick, entity, timeState)
	}
}

// performActivity executes the effects of performing an activity
func (br *BioRhythm) performActivity(activity ActivityType, tick int, entity *Entity, timeState TimeState) {
	activityState := br.Activities[activity]

	switch activity {
	case ActivitySleep:
		// Restore energy and reduce sleep need
		if entity.Energy < 100 {
			entity.Energy += 0.8 // Restore energy while sleeping
		}
		activityState.NeedLevel -= 0.05
		activityState.LastPerformed = tick
		br.EnergyAtLastSleep = entity.Energy

		// Complete sleep cycle after duration
		if activityState.Duration > 20 { // 20 ticks for a sleep cycle
			br.SleepCycles++
		}

	case ActivityEat:
		// Reduce hunger need (actual eating happens in interaction system)
		activityState.NeedLevel -= 0.03
		activityState.LastPerformed = tick
		br.LastMealTick = tick

	case ActivityDrink:
		// Reduce thirst need and slightly restore energy
		activityState.NeedLevel -= 0.04
		activityState.LastPerformed = tick
		br.LastDrinkTick = tick
		if entity.Energy < 100 {
			entity.Energy += 0.2
		}

	case ActivityPlay:
		// Reduce play need, costs energy but provides other benefits
		activityState.NeedLevel -= 0.02
		activityState.LastPerformed = tick
		entity.Energy -= 0.1 // Playing costs energy
		// Play can improve social traits over time (small chance)
		if rand.Float64() < 0.001 && entity.GetTrait("cooperation") < 1.0 {
			currentCoop := entity.GetTrait("cooperation")
			entity.SetTrait("cooperation", currentCoop+0.01)
		}

	case ActivityExplore:
		// Reduce exploration need, costs energy, may discover food
		activityState.NeedLevel -= 0.02
		activityState.LastPerformed = tick
		entity.Energy -= 0.15 // Exploring costs energy
		// Small chance to "discover" food (increase food-finding ability)
		if rand.Float64() < 0.002 {
			entity.Energy += 5.0 // Found something useful
		}

	case ActivityScavenge:
		// Reduce scavenging need (actual scavenging happens in interaction system)
		activityState.NeedLevel -= 0.025
		activityState.LastPerformed = tick

	case ActivityRest:
		// Restore small amount of energy and reduce rest need
		activityState.NeedLevel -= 0.03
		activityState.LastPerformed = tick
		if entity.Energy < 100 {
			entity.Energy += 0.3
		}

	case ActivitySocialize:
		// Reduce social need, small energy cost
		activityState.NeedLevel -= 0.02
		activityState.LastPerformed = tick
		entity.Energy -= 0.05
		// Socializing might improve cooperation
		if rand.Float64() < 0.001 && entity.GetTrait("cooperation") < 1.0 {
			currentCoop := entity.GetTrait("cooperation")
			entity.SetTrait("cooperation", currentCoop+0.005)
		}
	}

	// Ensure need levels don't go below 0
	if activityState.NeedLevel < 0 {
		activityState.NeedLevel = 0
	}
}

// GetCurrentActivity returns the activity the entity is currently performing
func (br *BioRhythm) GetCurrentActivity() ActivityType {
	for activity, state := range br.Activities {
		if state.IsActive {
			return activity
		}
	}
	return ActivityRest // Default
}

// GetActivityNeed returns the current need level for a specific activity
func (br *BioRhythm) GetActivityNeed(activity ActivityType) float64 {
	if state, exists := br.Activities[activity]; exists {
		return state.NeedLevel
	}
	return 0.0
}

// IsActivityTime checks if it's a good time for a specific activity based on schedule
func (br *BioRhythm) IsActivityTime(activity ActivityType, timeOfDay TimeOfDay) bool {
	scheduledActivities := br.ActivitySchedule[timeOfDay]
	for _, scheduledActivity := range scheduledActivities {
		if scheduledActivity == activity {
			return true
		}
	}
	return false
}

// GetActivityModifier returns an activity modifier based on circadian rhythm and current activity
func (br *BioRhythm) GetActivityModifier(entity *Entity, timeState TimeState) float64 {
	currentActivity := br.GetCurrentActivity()
	circadianPref := entity.GetTrait("circadian_preference")

	// Base modifier from circadian preference and time
	modifier := 1.0

	// Nocturnal entities get bonus at night, penalty during day
	if circadianPref < -0.3 && timeState.IsNight() {
		modifier += 0.3
	} else if circadianPref < -0.3 && !timeState.IsNight() {
		modifier -= 0.2
	}

	// Diurnal entities get bonus during day, penalty at night
	if circadianPref > 0.3 && !timeState.IsNight() {
		modifier += 0.3
	} else if circadianPref > 0.3 && timeState.IsNight() {
		modifier -= 0.2
	}

	// Activity-specific modifiers
	switch currentActivity {
	case ActivitySleep:
		modifier *= 0.1 // Very low activity when sleeping
	case ActivityRest:
		modifier *= 0.4 // Low activity when resting
	case ActivityPlay, ActivityExplore:
		modifier *= 1.2 // Higher activity when active
	case ActivityEat, ActivityDrink, ActivityScavenge:
		modifier *= 0.8 // Focused but lower movement
	default:
		modifier *= 1.0 // Normal activity
	}

	return math.Max(0.0, modifier)
}

// String returns a string representation of the biorhythm state
func (br *BioRhythm) String() string {
	currentActivity := br.GetCurrentActivity()
	activityNames := map[ActivityType]string{
		ActivitySleep:     "Sleep",
		ActivityEat:       "Eat",
		ActivityDrink:     "Drink",
		ActivityPlay:      "Play",
		ActivityExplore:   "Explore",
		ActivityScavenge:  "Scavenge",
		ActivityRest:      "Rest",
		ActivitySocialize: "Socialize",
	}

	if name, exists := activityNames[currentActivity]; exists {
		return name
	}
	return "Unknown"
}
