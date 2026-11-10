// Package model holds the data types of the SuMate (东坡嘴替) annotation
// system.  Everything here is JSON-friendly so that the Wails frontend can
// consume it directly and so the annotation data can live in plain JSON
// files that are diff-friendly and reviewable.
package model

// Strength grades the intensity of an emotional tag.
type Strength string

const (
	Slight   Strength = "slight"   // 轻微
	Moderate Strength = "moderate" // 中等
	Obvious  Strength = "obvious"  // 明显
	Profound Strength = "profound" // 深重
)

// Rank orders the strengths so they can be compared and summed.
func (s Strength) Rank() int {
	switch s {
	case Slight:
		return 1
	case Moderate:
		return 2
	case Obvious:
		return 3
	case Profound:
		return 4
	}
	return 0
}

// AtLeast reports whether s is at least as strong as other.
func (s Strength) AtLeast(other Strength) bool { return s.Rank() >= other.Rank() }

// Polarity is the overall emotional colour of a verse.
type Polarity string

const (
	Positive  Polarity = "positive"  // 正面
	Neutral   Polarity = "neutral"   // 中性
	Melanchol Polarity = "melanchol" // 悲怆/疗愈
	Mixed     Polarity = "mixed"     // 复合
)

// RootKey identifies one of the five top-level tag dimensions.
type RootKey string

const (
	RootMood     RootKey = "mood"     // 当前心境与内在情绪
	RootPlace    RootKey = "place"    // 当前地点与周遭环境
	RootWeather  RootKey = "weather"  // 当前时辰、节令与天气
	RootRelation RootKey = "relation" // 人际关系与社交状态
	RootOccasion RootKey = "occasion" // 人生境遇与现实处境
)

// RootOrder fixes the display order of the five roots.
var RootOrder = []RootKey{RootMood, RootPlace, RootWeather, RootRelation, RootOccasion}

// LifeStage is a period of Su Dongpo's life, used to frame the historical
// background of a verse.
type LifeStage string

const (
	StageEarlyCareer    LifeStage = "early_career"    // 早年科第与初入仕途
	StageInOffice       LifeStage = "in_office"       // 宦海升沉、顺境得志
	StageReformConflict LifeStage = "reform_conflict" // 新旧党争与外任
	StageHuangzhou      LifeStage = "huangzhou"       // 黄州贬谪
	StageReturn         LifeStage = "return"          // 哲宗朝复起
	StageHuizhou        LifeStage = "huizhou"         // 惠州贬谪
	StageDanzhou        LifeStage = "danzhou"         // 儋州贬谪
	StageLeisure        LifeStage = "leisure"         // 退隐闲居
	StageTravel         LifeStage = "travel"          // 羁旅行役
	StageLateYears      LifeStage = "late_years"      // 晚年归途与彻悟
)

// Era is the Chinese reign-era of a verse.
type Era string

// Reign eras used by Su Dongpo, in chronological order.
var EraOrder = []Era{
	"治平", "熙宁", "元丰", "元祐", "绍圣", "元符", "建中靖国", "崇宁",
}

// Emotion is a tagged emotional dimension with a graded strength.
type Emotion struct {
	// ID is the stable machine key, e.g. "melancholy.light_loneliness".
	ID string `json:"id"`
	// Name is the Chinese display label.
	Name string `json:"name"`
	// Dimension is one of positive / neutral / melanchol / dongpo.
	Dimension string `json:"dimension"`
	// Strength grades how strongly this emotion applies.
	Strength Strength `json:"strength"`
}

// Tangent is one of Su Dongpo's exclusive spiritual temperaments.
type Tangent string

const (
	TangentSorrowTranscend  Tangent = "sorrow_with_transcendence"        // 悲而能超
	TangentHardshipOptimism Tangent = "hardship_with_optimism"           // 苦中作乐
	TangentReconcile        Tangent = "reconciliation_with_imperfection" // 与不完美和解
	TangentTolerance        Tangent = "open_minded_tolerance"            // 挫败后的豁达宽容
	TangentDetach           Tangent = "freedom_from_worldly_obsession"   // 超越功名执念
	TangentAcceptFate       Tangent = "peaceful_acceptance_of_fate"      // 平静接受命运
)

// TangentOrder fixes the canonical order of the six Dongpo temperaments.
var TangentOrder = []Tangent{
	TangentSorrowTranscend,
	TangentHardshipOptimism,
	TangentReconcile,
	TangentTolerance,
	TangentDetach,
	TangentAcceptFate,
}

// TangentName maps a tangent to its Chinese display name.
var TangentName = map[Tangent]string{
	TangentSorrowTranscend:  "悲而能超",
	TangentHardshipOptimism: "苦中作乐",
	TangentReconcile:        "与不完美和解",
	TangentTolerance:        "挫败后的豁达宽容",
	TangentDetach:           "超越功名执念",
	TangentAcceptFate:       "平静接受命运",
}

// Background describes the historical context of a verse.
type Background struct {
	// Year is the Chinese reign year, e.g. "元丰五年"; empty if unknown.
	Year string `json:"year"`
	// GregorianYear is the approximate western year, 0 if unknown.
	GregorianYear int `json:"gregorian_year"`
	// Era is the reign era.
	Era Era `json:"era"`
	// Place is the modern place where the verse was composed.
	Place string `json:"place"`
	// Stage is the life stage of Su Dongpo.
	Stage LifeStage `json:"stage"`
	// Historical is a prose description of the circumstances.
	Historical string `json:"historical"`
	// Source is the 词题/小序 the facts were drawn from.
	Source string `json:"source"`
}

// Annotation is the full-dimensional annotation of one minimal verse unit.
type Annotation struct {
	// Literal is the surface meaning of the verse.
	Literal string `json:"literal"`
	// Conception is the extended artistic conception.
	Conception string `json:"conception"`
	// Philosophy is the hidden philosophical connotation.
	Philosophy string `json:"philosophy"`
	// Attitude is the implied attitude toward life.
	Attitude string `json:"attitude"`
	// Polarity is the overall emotional colour.
	Polarity Polarity `json:"polarity"`
	// Emotions are the graded emotional tags.
	Emotions []Emotion `json:"emotions"`
	// Tangents are the Dongpo-specific spiritual temperaments.
	Tangents []Tangent `json:"tangents"`
	// Crowd is the applicable audience (适用人群).
	Crowd string `json:"crowd"`
	// MoodThreshold is the mood level at which the verse applies.
	MoodThreshold string `json:"mood_threshold"`
	// ScenarioBoundary is the boundary of the applicable situation.
	ScenarioBoundary string `json:"scenario_boundary"`
	// Healing is the Dongpo-style healing attribute.
	Healing string `json:"healing"`
	// Concepts are the semantic concept keys used for matching.
	Concepts []string `json:"concepts"`
	// Keywords are surface keywords, used for display and explanation.
	Keywords []string `json:"keywords"`
	// Tags maps a root dimension to the leaf tag ids of that dimension.
	Tags map[RootKey][]string `json:"tags"`
}

// Verse is one minimal matching unit: a single classic line or a
// symmetrical couplet.
type Verse struct {
	// ID is the stable identifier, e.g. "wdtg-0007-2".
	ID string `json:"id"`
	// PoemID links back to the source poem.
	PoemID string `json:"poem_id"`
	// Tune is the 词牌.
	Tune string `json:"tune"`
	// Title is the 词题 / 小序.
	Title string `json:"title"`
	// Text is the verse itself, with modern punctuation.
	Text string `json:"text"`
	// Kind is "line" for a single line or "couplet" for a symmetrical pair.
	Kind string `json:"kind"`
	// Order is the position of the unit inside the poem.
	Order int `json:"order"`
	// Stanza is the index of the 片 the unit comes from.
	Stanza int `json:"stanza"`
	// Background is the historical framing.
	Background Background `json:"background"`
	// Annotation is the full semantic annotation.
	Annotation Annotation `json:"annotation"`
	// Tags maps a root key to the leaf tag ids of that dimension.
	Tags map[RootKey][]string `json:"tags"`
	// Score is the cached matching score of the verse.
	Score float64 `json:"score"`
}

// Poem is a source poem of the corpus.
type Poem struct {
	ID      string    `json:"id"`
	Tune    string    `json:"tune"`
	Title   string    `json:"title"`
	Text    string    `json:"text"`
	Stanzas []string  `json:"stanzas"`
	Sents   []string  `json:"sentences"`
	Kind    string    `json:"kind"`
	Quotes  []string  `json:"quotes"`
	Lacunae int       `json:"lacunae"`
	HasText bool      `json:"has_text"`
	Era     Era       `json:"era"`
	Year    string    `json:"year"`
	Stage   LifeStage `json:"stage"`
	Place   string    `json:"place"`
	Order   int       `json:"order"`
}

// EraYear pairs a reign era with a western year.  It is used to resolve the
// cyclical year names ("丙辰") that appear in 词题.
type EraYear struct {
	Era  Era
	Year int
}
