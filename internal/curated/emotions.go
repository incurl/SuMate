package curated

import "github.com/sumate/sumate/internal/model"

// EmotionDef describes one emotional dimension of the annotation system.
type EmotionDef struct {
	// Key is the stable machine key.
	Key string `json:"key"`
	// Name is the Chinese display name.
	Name string `json:"name"`
	// Dimension groups the emotion: positive / neutral / melanchol / dongpo.
	Dimension string `json:"dimension"`
	// Parent groups sub-emotions under a broader feeling.
	Parent string `json:"parent"`
}

// EmotionDimensions are the four emotional dimensions of the system.
var EmotionDimensions = []string{"positive", "neutral", "melanchol", "dongpo"}

// emotionList is the full emotion lexicon.
var emotionList = []EmotionDef{
	// ---------------------------------------------------------- 正面情绪
	{Key: "positive.serene_joy", Name: "恬静之喜", Dimension: "positive", Parent: "positive"},
	{Key: "positive.casual_leisure", Name: "闲适从容", Dimension: "positive", Parent: "positive"},
	{Key: "positive.unrestrained_freedom", Name: "肆意洒脱", Dimension: "positive", Parent: "positive"},
	{Key: "positive.open_minded_magnanimity", Name: "豁达大度", Dimension: "positive", Parent: "positive"},
	{Key: "positive.relieved_mood", Name: "如释重负", Dimension: "positive", Parent: "positive"},
	{Key: "positive.peaceful_tranquility", Name: "安宁平静", Dimension: "positive", Parent: "positive"},
	{Key: "positive.hopeful_outlook", Name: "仍有希望", Dimension: "positive", Parent: "positive"},
	{Key: "positive.warm_gentleness", Name: "温柔亲和", Dimension: "positive", Parent: "positive"},
	{Key: "positive.romantic_relaxation", Name: "风流闲适", Dimension: "positive", Parent: "positive"},
	{Key: "positive.worldly_ease", Name: "人间烟火的从容", Dimension: "positive", Parent: "positive"},
	{Key: "positive.plain_contentment", Name: "平淡知足", Dimension: "positive", Parent: "positive"},

	// ---------------------------------------------------------- 中性哲思
	{Key: "neutral.life_insight", Name: "人生洞见", Dimension: "neutral", Parent: "neutral"},
	{Key: "neutral.world_observation", Name: "世事观察", Dimension: "neutral", Parent: "neutral"},
	{Key: "neutral.calm_detachment", Name: "冷静超脱", Dimension: "neutral", Parent: "neutral"},
	{Key: "neutral.rational_thinking", Name: "理性思考", Dimension: "neutral", Parent: "neutral"},
	{Key: "neutral.plain_contentment", Name: "平淡知足", Dimension: "neutral", Parent: "neutral"},
	{Key: "neutral.indifferent_perspective", Name: "不萦于物", Dimension: "neutral", Parent: "neutral"},
	{Key: "neutral.tranquil_solitude", Name: "宁静独处", Dimension: "neutral", Parent: "neutral"},

	// ---------------------------------------------------------- 低回疗愈
	{Key: "melanchol.light_loneliness", Name: "轻微孤独", Dimension: "melanchol", Parent: "melanchol"},
	{Key: "melanchol.gentle_nostalgia", Name: "淡淡怀旧", Dimension: "melanchol", Parent: "melanchol"},
	{Key: "melanchol.distant_longing", Name: "遥远想念", Dimension: "melanchol", Parent: "melanchol"},
	{Key: "melanchol.subtle_frustration", Name: "淡淡的窝心", Dimension: "melanchol", Parent: "melanchol"},
	{Key: "melanchol.career_melancholy", Name: "仕途之叹", Dimension: "melanchol", Parent: "melanchol"},
	{Key: "melanchol.parting_sadness", Name: "离别之悲", Dimension: "melanchol", Parent: "melanchol"},
	{Key: "melanchol.life_helplessness", Name: "人生无奈", Dimension: "melanchol", Parent: "melanchol"},
	{Key: "melanchol.trivial_weariness", Name: "琐碎疲惫", Dimension: "melanchol", Parent: "melanchol"},

	// ---------------------------------------------------------- 东坡气质
	{Key: "dongpo.sorrow_with_transcendence", Name: "悲而能超", Dimension: "dongpo", Parent: "dongpo"},
	{Key: "dongpo.hardship_with_optimism", Name: "苦中作乐", Dimension: "dongpo", Parent: "dongpo"},
	{Key: "dongpo.reconciliation_with_imperfection", Name: "与不完美和解", Dimension: "dongpo", Parent: "dongpo"},
	{Key: "dongpo.open_minded_tolerance", Name: "挫败后的豁达", Dimension: "dongpo", Parent: "dongpo"},
	{Key: "dongpo.freedom_from_worldly_obsession", Name: "不萦于物", Dimension: "dongpo", Parent: "dongpo"},
	{Key: "dongpo.peaceful_acceptance_of_fate", Name: "平静接受命运", Dimension: "dongpo", Parent: "dongpo"},
}

// Emotions returns the emotion lexicon keyed by emotion key.
func Emotions() map[string]EmotionDef {
	m := make(map[string]EmotionDef, len(emotionList))
	for _, e := range emotionList {
		m[e.Key] = e
	}
	return m
}

// curatedEmotionKeys is the set of known emotion keys.
func curatedEmotionKeys() map[string]bool {
	m := map[string]bool{}
	for _, e := range emotionList {
		m[e.Key] = true
	}
	return m
}

// knownTangent reports whether the tangent key is one of the six Dongpo
// spiritual temperaments.
func knownTangent(k string) bool {
	for _, t := range model.TangentOrder {
		if string(t) == k {
			return true
		}
	}
	return false
}
