package curated

// ModernTrigger is a modern colloquial word that names the same state as the
// concept's classical trigger words.
//
// This file exists because of a measured failure.  The free-text field was
// almost inert: 「我想一个人待着」, 「我很累不想动」 and 「思念远方的人」
// all returned the same verse.  The cause was not a lack of semantic power
// -- the lexicon already expressed all of those states (8/8) -- but a
// vocabulary gap: every one of the 936 trigger words was classical, and
// 0 of the 60 words a modern reader actually types (累 / 困 / 想家 / 难过)
// appeared in the lexicon.
//
// Adding these words closes the gap for the cost of a few dozen strings.
// It also keeps the match *explainable*: the user can be told that
// 「疲惫」 selected these verses, which a 200 MB embedding model could never
// say.
//
// The words are grouped by the concept they belong to, so the mapping stays
// auditable: every modern word points at exactly one concept.
var modernTriggers = map[string][]string{
	// 一人、独自 —— 独处
	"serene.joy.quiet": {
		"一个人", "独自", "自己待着", "待着", "没人管", "自己一个",
		"开心", "高兴", "快乐", "愉快", "舒心", "轻快", "美滋滋",
	},
	"moon.solitude": {
		"孤独", "寂寞", "一个人", "落单", "空落落", "冷清", "没人",
	},

	// 疲、倦 —— 疲惫与无所事事
	"idleness.day": {
		"累", "疲劳", "疲", "倦", "疲乏", "没劲", "无力", "不想动",
		"提不起劲", "躺着", "歇着", "不想说话",
	},
	"relief.after_storm": {
		"放松", "缓过来", "缓", "平复", "缓一缓",
	},

	// 雨 —— 天气
	"rain.drizzle": {
		"下雨", "阴雨", "潮湿", "湿", "闷", "雨下",
	},
	"wind": {
		"刮风", "起风", "大风",
	},
	"snow": {
		"下雪", "雪天", "好冷",
	},

	// 思念 —— 人
	"human.love": {
		"想他", "想她", "想念", "惦记", "远方的", "想见的",
	},
	"moon.longing": {
		"想念", "惦记", "远方", "睡不着想", "陪着我",
	},
	"human.mother": {
		"想妈", "想娘", "我妈", "我妈想",
	},
	"human.brother": {
		"想弟弟", "想哥哥", "兄弟", "手足",
	},

	// 怀旧 —— 处境
	"nostalgia.old_place": {
		"想家", "回家", "老家", "故乡", "从前", "以前", "当年",
		"回忆", "以前的事", "老朋友",
	},
	"nostalgia.time": {
		"一晃", "转眼", "白驹过隙", "日子过得",
	},

	// 低落 —— 心境
	"life.transience": {
		"难过", "低落", "郁闷", "不开心", "emo", "心沉", "堵得慌",
		"闷闷的", "提不起劲", "没意思",
	},
	"tangled.ambivalent": {
		"纠结", "又爱又恨", "放不下", "不甘心又舍不得",
	},

	// 被压着 —— 处境
	"occasion.life.illness": {
		"压力", "喘不过气", "扛不住", "撑不住", "被逼", "窒息",
		"透不过气", "绷着",
	},
	"occasion.career.passed": {
		"不被看重", "没人用我", "白干", "被打压", "受委屈",
	},

	// 平淡 —— 心境
	"observe.world": {
		"平淡", "平常", "没什么", "一般", "看开了", "就这样",
	},
	"life.plain": {
		"平淡", "平常心", "凑合", "将就", "粗茶淡饭",
	},

	// 无奈 —— 心境
	"helpless.against_fate": {
		"不知道怎么办", "没主意", "茫然", "迷茫", "该何去何从",
		"无奈", "怎么办", "不知如何", "两难",
	},
	"frustration.unfulfilled": {
		"不甘心", "白费", "不服", "窝火", "憋屈",
	},
	"tangled.words": {
		"说不出口", "开不了口", "憋着", "没脸说", "难以启齿",
	},

	// 轻快 —— 心境
	"romantic.flirt": {
		"风流", "好看", "美", "心动",
	},
	"daily.food": {
		"好吃", "想吃点", "饿了", "嘴馋",
	},
	"daily.sleep": {
		"睡不着", "失眠", "熬夜", "困了", "睡不好",
	},
	"worldly.grounded": {
		"过日子", "烟火气", "踏实", "安心",
	},
}

// modernByConcept returns the modern trigger words of every concept.
func modernByConcept() map[string][]string {
	m := map[string][]string{}
	for k, ws := range modernTriggers {
		if len(ws) == 0 {
			continue
		}
		m[k] = ws
	}
	return m
}
