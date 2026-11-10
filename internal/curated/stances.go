package curated

// Stance is a philosophical position that a verse takes about how to live.
//
// A stance is neither a mood nor a temperament.  A temperament says *how*
// Su Dongpo bears what happens to him ("悲而能超"); a stance says *what he
// claims* about the world ("世事需有人担当；空谈心性是逃").
//
// The philosophical content of a stance is its Premise and its Shadow:
//
//   - the Premise is what must be true of the reader for the stance to
//     *help* rather than merely sound wise;
//   - the Shadow is what the stance becomes when the premise fails.
//
// This is what separates a mouthpiece from a quotation.  「此心安处是吾乡」
// is consolation to a man who has been exiled and can find no other footing;
// to a man who still has a road to travel it is an excuse to stop walking.
// The same words land differently depending on where the reader stands.
//
// A stance also records where it was *earned*.  Su Dongpo's most quoted
// sentences were written by a man who had already been through something.
// 「拣尽寒枝不肯栖」 is not a piece of advice; it is a report from a winter
// night.  The verse carries its own conditions.
type Stance struct {
	// Key is the stable machine key.
	Key string `json:"key"`
	// Name is the Chinese display name.
	Name string `json:"name"`
	// Claim is the proposition the stance asserts.
	Claim string `json:"claim"`
	// Premise is what must hold for the stance to help.
	Premise string `json:"premise"`
	// Shadow is what the stance turns into when the premise fails.
	Shadow string `json:"shadow"`
	// Earned describes the condition under which Su Dongpo could say it.
	Earned string `json:"earned"`
	// Supports are the tag ids whose presence in the reader's situation
	// suggests the premise is met.
	Supports []string `json:"supports"`
	// Undermines are the tag ids whose presence suggests the premise has
	// failed.
	Undermines []string `json:"undermines"`
	// Tangents are the Dongpo temperaments that express this stance.
	Tangents []string `json:"tangents"`
	// Concepts are the concept keys that carry this stance.
	Concepts []string `json:"concepts"`
}

// StanceKeys returns the keys of the stances that a concept carries.
//
// The mapping goes from concept to stance rather than the other way round,
// because the concept is what the keyword matcher produces: when a verse
// matches 「此心安处是吾乡」 it matches the concept `occasion.exile_joy`,
// and the stance follows from there.
func StanceKeys() map[string][]string {
	m := map[string][]string{}
	for _, s := range stanceList {
		for _, ck := range s.Concepts {
			m[ck] = append(m[ck], s.Key)
		}
	}
	return m
}

// Stances returns the philosophical stance lexicon.
func Stances() map[string]Stance {
	m := make(map[string]Stance, len(stanceList))
	for _, s := range stanceList {
		m[s.Key] = s
	}
	return m
}

// StanceOrder fixes the canonical order of the stances.
var StanceOrder = []string{
	"suffering.perspective",
	"suffering.acceptance",
	"impermanence.attachment",
	"fate.limits",
	"fate.trust",
	"effort.worth",
	"effort.ease",
	"self.with_world",
	"self.principle",
	"joy.notice",
	"meaning.participation",
}

// stanceList is the canonical philosophical stance lexicon.
//
// The stances are ordered from the most consoling to the most demanding,
// because that is the order in which a reader usually needs them: first the
// reframe, then the demand.
var stanceList = []Stance{
	// ------------------------------------------------------------ 苦难
	{
		Key:     "suffering.perspective",
		Name:    "换个高度看同一件事",
		Claim:   "你之所以这么苦，一部分是因为你站得太近。",
		Premise: "这件事已经无法改变，而你还有余裕去站得更高一点。",
		Shadow:  "如果事情其实还可以改变，那么「想开一点」就只是不去面对。",
		Earned:  "苏东坡写这句时刚经历贬谪；他见过不站高一点会变成什么。",
		Supports: []string{
			"mood.insight.zen", "mood.insight.observe", "place.nature.mountain",
			"place.city.tower", "mood.joy.casual",
		},
		Undermines: []string{
			"occasion.exile.hardship.poverty", "relation.official.rival",
			"mood.low.helpless", "occasion.life.illness",
		},
		Tangents: []string{"freedom_from_worldly_obsession", "open_minded_tolerance"},
		Concepts: []string{"mountain.high", "life.transience", "observe.world"},
	},
	{
		Key:     "suffering.acceptance",
		Name:    "苦已经过去，不必再证明",
		Claim:   "熬过来了的人，不需要反复讲述自己受过多少苦。",
		Premise: "最难的阶段已经过去，现在的难受是余波而不是新伤。",
		Shadow:  "如果苦还在继续，这句话就成了「不许喊疼」。",
		Earned:  "苏轼从黄州回来之后才写得出来；此前他还在痛里。",
		Supports: []string{
			"mood.joy.relief", "occasion.exile.grand", "occasion.career.appointed",
			"mood.low.weary",
		},
		Undermines: []string{
			"occasion.exile.hardship", "mood.low.helpless", "occasion.life.illness",
			"relation.official.rival",
		},
		Tangents: []string{"hardship_with_optimism", "peaceful_acceptance_of_fate"},
		Concepts: []string{"relief.after_storm", "life.plain", "occasion.official_return"},
	},
	// ------------------------------------------------------------ 聚散
	{
		Key:     "impermanence.attachment",
		Name:    "抓不住的东西，本来就抓不住",
		Claim:   "万物都在流，再用力也留不住；正因留不住，人才会认真。",
		Premise: "你珍视的东西确实是有限的，而你知道它会过去。",
		Shadow:  "如果一切都会过去，这句话也可以变成不必珍惜的借口。",
		Earned:  "东坡一生送别，写的不是别离本身，是别离之后仍然要活。",
		Supports: []string{
			"mood.low.nostalgia", "mood.low.longing", "weather.sky.moon",
			"relation.family", "relation.friend.old",
		},
		Undermines: []string{
			"mood.joy.release", "mood.insight.zen.detached", "occasion.daily.idle",
		},
		Tangents: []string{"sorrow_with_transcendence", "peaceful_acceptance_of_fate"},
		Concepts: []string{"river.flow", "nostalgia.time", "human.reunion"},
	},
	// ------------------------------------------------------------ 命运
	{
		Key:     "fate.limits",
		Name:    "有些事不是你不努力",
		Claim:   "人的能力有边界，意志不能无限延长它。",
		Premise: "这件事确实超出了你所能控制的范围。",
		Shadow:  "「不是你的错」也可以用来掩盖「你没试」。",
		Earned:  "苏轼三贬三起，知道什么是真的无力。",
		Supports: []string{
			"mood.low.helpless", "occasion.exile.hardship.solitude",
			"mood.low.tangled", "relation.society.excluded",
		},
		Undermines: []string{
			"occasion.career.appointed", "mood.joy.hope", "mood.joy.release",
		},
		Tangents: []string{"peaceful_acceptance_of_fate", "sorrow_with_transcendence"},
		Concepts: []string{"helpless.against_fate", "occasion.lost", "life.futility"},
	},
	{
		Key:     "fate.trust",
		Name:    "交给不可控的那部分",
		Claim:   "把自己这一段走好，剩下的是另一回事，不必现在结算。",
		Premise: "你已经做了你能做的，剩下的确实不由你决定。",
		Shadow:  "如果根本没做，「交给命运」就是拖延的漂亮说法。",
		Earned:  "东坡说这话时正在等一纸判决，闲笔写的是自己。",
		Supports: []string{
			"mood.insight.zen", "mood.joy.casual", "occasion.journey",
			"mood.low.weary",
		},
		Undermines: []string{
			"mood.low.frustration", "occasion.career.passed", "mood.joy.release",
		},
		Tangents: []string{"peaceful_acceptance_of_fate", "freedom_from_worldly_obsession"},
		Concepts: []string{"life.river_acceptance", "river.breeze", "post.inn"},
	},
	// ------------------------------------------------------------ 进取
	{
		Key:     "effort.worth",
		Name:    "不因为无用而放弃理想",
		Claim:   "现在做不成，不代表不该想；理想的火可以换个地方烧。",
		Premise: "你想要的东西本身是对的，只是眼下这条路走不通。",
		Shadow:  "如果那个理想本来就不该是你的，「不放弃」只是执迷。",
		Earned:  "苏轼一生想做事而不得做，直到最后也没有真正放下。",
		Supports: []string{
			"mood.low.frustration", "occasion.career.passed", "mood.joy.hope",
			"relation.official.patron",
		},
		Undermines: []string{
			"mood.insight.zen.detached", "mood.low.helpless", "mood.joy.casual",
		},
		Tangents: []string{"hardship_with_optimism", "sorrow_with_transcendence"},
		Concepts: []string{"frustration.unfulfilled", "ambition.unjut", "ambition.history"},
	},
	{
		Key:     "effort.ease",
		Name:    "把用力松开一点",
		Claim:   "不是所有事都值得用全力，有些事轻一点反而做得久。",
		Premise: "你正在做的事里有相当一部分其实并不重要。",
		Shadow:  "如果全都很重要，「别太用力」就是逃避。",
		Earned:  "黄州的粗茶淡饭让东坡发现，被贬之后唯一还能自己决定的是吃不吃得下饭。",
		Supports: []string{
			"mood.joy.casual", "mood.joy.casual.idle", "occasion.daily.idle",
			"mood.joy.relief",
		},
		Undermines: []string{
			"mood.low.frustration", "occasion.career.passed", "mood.joy.hope",
		},
		Tangents: []string{"reconciliation_with_imperfection", "freedom_from_worldly_obsession"},
		Concepts: []string{"idleness.day", "life.plain", "worldly.grounded"},
	},
	// ------------------------------------------------------------ 自处
	{
		Key:     "self.with_world",
		Name:    "与天地相接，而不是与人争",
		Claim:   "把自己放回天地之间，很多问题会自己变小。",
		Premise: "你在意的那件事，用人的尺度看是巨大的。",
		Shadow:  "如果那件事真的会伤到你，「天地很大」是廉价的安慰。",
		Earned:  "东坡写赤壁怀古时刚从黄州回来不久，他见过小。",
		Supports: []string{
			"place.nature.river", "place.nature.mountain", "mood.insight.zen",
			"mood.insight.zen.detached",
		},
		Undermines: []string{
			"relation.romance.heartbreak", "occasion.life.death.parent",
			"mood.low.helpless",
		},
		Tangents: []string{"freedom_from_worldly_obsession", "sorrow_with_transcendence"},
		Concepts: []string{"river.flow", "mountain.high", "ambition.history"},
	},
	{
		Key:     "self.principle",
		Name:    "守住自己那一条",
		Claim:   "处境可以改，标准不必改；不改标准，就还没输。",
		Premise: "你心里确实有一条不肯让的线，而且你知道它是什么。",
		Shadow:  "如果那条线其实是为了面子，「不改」就只是不肯认输。",
		Earned:  "东坡在黄州写诗，在惠州煮肉，在儋州教书——环境一再变，他没变。",
		Supports: []string{
			"lotus.purity", "mood.insight.zen", "occasion.exile.hardship",
			"relation.official.rival",
		},
		Undermines: []string{
			"mood.joy.casual", "mood.joy.relief", "mood.low.helpless",
		},
		Tangents: []string{"open_minded_tolerance", "reconciliation_with_imperfection"},
		Concepts: []string{"lotus.purity", "plum.bloom", "rival.office"},
	},
	// ------------------------------------------------------------ 日常
	{
		Key:     "joy.notice",
		Name:    "清欢是可以自己造出来的",
		Claim:   "不必等到什么大事发生才能高兴；此刻就有值得高兴的。",
		Premise: "你已经愿意停下来看一看了。",
		Shadow:  "如果是被迫麻木，「清欢」就成了不去感受的借口。",
		Earned:  "东坡写「人间有味是清欢」，那年他刚被贬，还没有清欢可写。",
		Supports: []string{
			"mood.joy.serene", "mood.joy.casual", "daily.food", "daily.music",
			"occasion.daily.meal",
		},
		Undermines: []string{
			"mood.low.helpless", "mood.low.frustration", "occasion.life.illness",
		},
		Tangents: []string{"hardship_with_optimism", "reconciliation_with_imperfection"},
		Concepts: []string{"serene.joy.quiet", "daily.food", "daily.music"},
	},
	{
		Key:     "meaning.participation",
		Name:    "意义在参与里，不在旁观里",
		Claim:   "看见世界的好与不好，都不能代替你真的去做点什么。",
		Premise: "你还有能承担的事，而且愿意承担。",
		Shadow:  "如果该承担的你并不打算承担，「参与」就成了一种姿态。",
		Earned:  "东坡被贬到最远的地方仍然教书写字；他的意义感从来没被剥夺过。",
		Supports: []string{
			"mood.joy.romantic", "occasion.life.renewal", "relation.official.patron",
			"mood.joy.hope",
		},
		Undermines: []string{
			"mood.insight.zen.detached", "mood.joy.casual.idle",
			"mood.low.helpless",
		},
		Tangents: []string{"open_minded_tolerance", "hardship_with_optimism"},
		Concepts: []string{"patron.friend", "ambition.unjut", "worldly.grounded"},
	},
}
