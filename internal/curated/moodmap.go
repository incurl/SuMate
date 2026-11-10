package curated

import "github.com/sumate/sumate/internal/model"

// MoodEmotion links a mood tag in the tree to the emotions it stands for.
//
// This is what lets the product match a *state of mind* the user picked in
// the tree against the emotional profile of a verse, without the user ever
// having to describe their feelings in words.
type MoodEmotion struct {
	// Tag is the mood tag id in the tree.
	Tag string
	// Emotions are the emotion keys this mood expresses.
	Emotions []string
	// Strength is the intensity the mood tag implies.
	Strength model.Strength
}

// MoodEmotionMap returns the mapping from mood tags to emotions.
func MoodEmotionMap() map[string]MoodEmotion {
	m := make(map[string]MoodEmotion, len(moodEmotionList))
	for _, x := range moodEmotionList {
		m[x.Tag] = x
	}
	return m
}

// moodEmotionList is the canonical mood-to-emotion mapping.
var moodEmotionList = []MoodEmotion{
	// ---- 喜悦与轻快
	{"mood.joy.serene", []string{"positive.serene_joy", "positive.peaceful_tranquility"}, model.Moderate},
	{"mood.joy.casual", []string{"positive.casual_leisure", "positive.worldly_ease"}, model.Obvious},
	{"mood.joy.casual.idle", []string{"positive.casual_leisure"}, model.Obvious},
	{"mood.joy.casual.food", []string{"positive.worldly_ease", "positive.casual_leisure"}, model.Obvious},
	{"mood.joy.release", []string{"positive.unrestrained_freedom"}, model.Obvious},
	{"mood.joy.release.drunk", []string{"positive.unrestrained_freedom", "positive.casual_leisure"}, model.Obvious},
	{"mood.joy.relief", []string{"positive.relieved_mood"}, model.Obvious},
	{"mood.joy.hope", []string{"positive.hopeful_outlook"}, model.Obvious},
	{"mood.joy.warm", []string{"positive.warm_gentleness"}, model.Obvious},
	{"mood.joy.worldly", []string{"positive.worldly_ease", "positive.plain_contentment"}, model.Moderate},
	{"mood.joy.romantic", []string{"positive.romantic_relaxation"}, model.Moderate},

	// ---- 通达与洞察
	{"mood.insight", []string{"neutral.life_insight"}, model.Moderate},
	{"mood.insight.zen", []string{"neutral.calm_detachment", "neutral.tranquil_solitude"}, model.Obvious},
	{"mood.insight.zen.detached", []string{"neutral.indifferent_perspective", "neutral.calm_detachment"}, model.Profound},
	{"mood.insight.rational", []string{"neutral.rational_thinking"}, model.Moderate},
	{"mood.insight.plain", []string{"neutral.plain_contentment", "positive.plain_contentment"}, model.Obvious},
	{"mood.insight.plain.enough", []string{"positive.plain_contentment"}, model.Profound},
	{"mood.insight.solitude", []string{"neutral.tranquil_solitude"}, model.Moderate},
	{"mood.insight.observe", []string{"neutral.world_observation"}, model.Moderate},

	// ---- 低回与疗愈
	{"mood.low.lonely", []string{"melanchol.light_loneliness"}, model.Moderate},
	{"mood.low.nostalgia", []string{"melanchol.gentle_nostalgia"}, model.Moderate},
	{"mood.low.longing", []string{"melanchol.distant_longing"}, model.Obvious},
	{"mood.low.frustration", []string{"melanchol.subtle_frustration", "melanchol.career_melancholy"}, model.Obvious},
	{"mood.low.career", []string{"melanchol.career_melancholy"}, model.Obvious},
	{"mood.low.parting", []string{"melanchol.parting_sadness"}, model.Obvious},
	{"mood.low.helpless", []string{"melanchol.life_helplessness"}, model.Profound},
	{"mood.low.weary", []string{"melanchol.trivial_weariness"}, model.Obvious},
	{"mood.low.tangled", []string{"melanchol.distant_longing", "melanchol.subtle_frustration"}, model.Obvious},
	{"mood.low.tangled.inexpressible", []string{"melanchol.life_helplessness", "melanchol.distant_longing"}, model.Profound},
	{"mood.low.tangled.ambivalent", []string{"melanchol.subtle_frustration"}, model.Obvious},
	{"mood.low.tangled.ashamed", []string{"melanchol.life_helplessness"}, model.Obvious},
}
