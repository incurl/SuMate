package tagtree

import "github.com/sumate/sumate/internal/model"

// spec is a compact description of one node in the canonical tag tree.
type spec struct {
	id    string
	label string
	hint  string
}

// Build assembles the canonical SuMate tag tree.
//
// The tree has five fixed roots and nests freely below them; the depth is
// intentionally uneven because some dimensions (mood, occasion) carry a much
// richer vocabulary than others (weather).
func Build() *Tree {
	t := New()

	add := func(root model.RootKey, parent string, s spec) {
		t.Add(model.TagNode{
			ID:     s.id,
			Root:   root,
			Parent: parent,
			Label:  s.label,
			Hint:   s.hint,
		})
	}

	// ---------------------------------------------------------------- 1. 心境
	r := model.RootMood
	add(r, "", spec{"mood", "当前心境与内在情绪", "此刻说不出口的那种感觉"})
	// -- 喜悦与轻快
	add(r, "mood", spec{"mood.joy", "喜悦与轻快", "由内而外的舒展"})
	add(r, "mood.joy", spec{"mood.joy.serene", "恬淡之喜", "不张扬的、底色般的喜悦"})
	add(r, "mood.joy", spec{"mood.joy.casual", "闲适从容", "把时间浪费在喜欢的事上"})
	add(r, "mood.joy", spec{"mood.joy.release", "肆意洒脱", "松开束缚后的自由"})
	add(r, "mood.joy", spec{"mood.joy.relief", "如释重负", "长久压力之后的松弛"})
	add(r, "mood.joy", spec{"mood.joy.hope", "仍有期待", "相信事情会好转"})
	add(r, "mood.joy", spec{"mood.joy.warm", "温柔亲近", "对人对事的柔软"})
	add(r, "mood.joy", spec{"mood.joy.worldly", "人间烟火的踏实", "知道活着是什么样子"})
	add(r, "mood.joy", spec{"mood.joy.romantic", "风流旖旎", "带着审美与情的愉悦"})
	// -- 通达与洞察
	add(r, "mood", spec{"mood.insight", "通达与洞察", "看明白之后的平静"})
	add(r, "mood.insight", spec{"mood.insight.zen", "超然物外", "不为外物所动"})
	add(r, "mood.insight", spec{"mood.insight.rational", "理性清明", "把事情拆开来看"})
	add(r, "mood.insight", spec{"mood.insight.plain", "平淡知足", "粗茶淡饭即是圆满"})
	add(r, "mood.insight", spec{"mood.insight.solitude", "宁静独处", "与己安静相处"})
	add(r, "mood.insight", spec{"mood.insight.observe", "冷眼观世", "看得很清楚却不涉入"})
	// -- 低回与疗愈
	add(r, "mood", spec{"mood.low", "低回与疗愈", "需要被慢慢安放的情绪"})
	add(r, "mood.low", spec{"mood.low.lonely", "孤独", "无人同行的安静"})
	add(r, "mood.low", spec{"mood.low.nostalgia", "怀旧", "旧时光、旧人、旧事"})
	add(r, "mood.low", spec{"mood.low.longing", "思念", "隔着距离与时间的心事"})
	add(r, "mood.low", spec{"mood.low.frustration", "壮志难酬", "想做却做不成"})
	add(r, "mood.low", spec{"mood.low.career", "仕途失意", "在仕途上受了委屈"})
	add(r, "mood.low", spec{"mood.low.parting", "离别", "送别与被送别"})
	add(r, "mood.low", spec{"mood.low.helpless", "无奈与无力", "有些事改不了"})
	add(r, "mood.low", spec{"mood.low.weary", "琐碎疲惫", "被日常磨得没了力气"})
	add(r, "mood.low", spec{"mood.low.tangled", "情绪纠缠难言", "说不清、又放不下"})

	// ---------------------------------------------------------------- 2. 地点
	r = model.RootPlace
	add(r, "", spec{"place", "当前地点与周遭环境", "你此刻站在哪里"})
	add(r, "place", spec{"place.nature", "山水自然", "天地之间"})
	add(r, "place.nature", spec{"place.nature.river", "江河湖海", "水面开阔处"})
	add(r, "place.nature", spec{"place.nature.mountain", "山岭", "登高与远望"})
	add(r, "place.nature", spec{"place.nature.forest", "林泉草木", "树影与泉声之间"})
	add(r, "place.nature", spec{"place.nature.field", "田园原野", "稻麦与村庄"})
	add(r, "place.nature", spec{"place.nature.garden", "园圃庭院", "自家庭院与园林"})
	add(r, "place", spec{"place.city", "城市街巷", "市井与楼台"})
	add(r, "place.city", spec{"place.city.tower", "高楼远眺", "登楼所见"})
	add(r, "place.city", spec{"place.city.street", "街市酒肆", "人群与烟火"})
	add(r, "place.city", spec{"place.city.temple", "寺观楼阁", "清幽之地"})
	add(r, "place.city", spec{"place.city.frontier", "边塞关外", "天涯与战场"})
	add(r, "place", spec{"place.indoor", "室内居所", "一灯一室之内"})
	add(r, "place.indoor", spec{"place.indoor.bed", "枕畔帐中", "睡意与梦境之间"})
	add(r, "place.indoor", spec{"place.indoor.window", "窗前", "推窗所见"})
	add(r, "place.indoor", spec{"place.indoor.hall", "厅堂席间", "宴饮与相见"})
	add(r, "place", spec{"place.travel", "旅途之中", "在路上"})
	add(r, "place.travel", spec{"place.travel.boat", "舟中", "船上与水上"})
	add(r, "place.travel", spec{"place.travel.post", "驿站客舍", "旅途的停顿"})
	add(r, "place", spec{"place.exile", "贬谪之所", "不在自己想去的地方"})

	// ---------------------------------------------------------------- 3. 时令
	r = model.RootWeather
	add(r, "", spec{"weather", "当前时辰、节令与天气", "此刻的天与时"})
	add(r, "weather", spec{"weather.season", "四季节令", "年轮中的位置"})
	add(r, "weather.season", spec{"weather.season.spring", "春", "万物初生"})
	add(r, "weather.season", spec{"weather.season.summer", "夏", "溽热与骤雨"})
	add(r, "weather.season", spec{"weather.season.autumn", "秋", "清高与萧瑟"})
	add(r, "weather.season", spec{"weather.season.winter", "冬", "寒意与围炉"})
	add(r, "weather.season", spec{"weather.season.dusk", "节令交关", "立春、重阳、除夕"})
	add(r, "weather", spec{"weather.sky", "天象", "抬头所见"})
	add(r, "weather.sky", spec{"weather.sky.moon", "月", "月色与月轮"})
	add(r, "weather.sky", spec{"weather.sky.sun", "日光", "日影与晴光"})
	add(r, "weather.sky", spec{"weather.sky.cloud", "云", "云卷云舒"})
	add(r, "weather.sky", spec{"weather.sky.star", "星辰", "星河与渔火"})
	add(r, "weather", spec{"weather.rain", "风雨", "雨、雪、风"})
	add(r, "weather.rain", spec{"weather.rain.drizzle", "微雨细雪", "润物无声的天气"})
	add(r, "weather.rain", spec{"weather.rain.storm", "风雨骤至", "骤雨与狂风"})
	add(r, "weather.rain", spec{"weather.rain.snow", "雪", "积雪与雪后"})
	add(r, "weather", spec{"weather.time", "时辰", "一天中的时刻"})
	add(r, "weather.time", spec{"weather.time.dawn", "清晨", "天光初开"})
	add(r, "weather.time", spec{"weather.time.noon", "正午", "日影当头"})
	add(r, "weather.time", spec{"weather.time.dusk", "黄昏", "日落与灯起"})
	add(r, "weather.time", spec{"weather.time.night", "深夜", "人静而思多"})
	add(r, "weather", spec{"weather.humidity", "冷暖干湿", "身体的感受"})
	add(r, "weather.humidity", spec{"weather.humidity.hot", "溽热", "汗与扇"})
	add(r, "weather.humidity", spec{"weather.humidity.cold", "清寒", "冷到骨头里"})
	add(r, "weather.humidity", spec{"weather.humidity.muggy", "潮湿", "雨后的黏腻"})

	// ---------------------------------------------------------------- 4. 人际
	r = model.RootRelation
	add(r, "", spec{"relation", "人际关系与社交状态", "你与谁之间的事"})
	add(r, "relation", spec{"relation.family", "家人至亲", "血缘与姻亲"})
	add(r, "relation.family", spec{"relation.family.brother", "兄弟", "手足之间"})
	add(r, "relation.family", spec{"relation.family.son", "子女", "为人父母"})
	add(r, "relation.family", spec{"relation.family.wife", "夫妻", "同室之人"})
	add(r, "relation.family", spec{"relation.family.parent", "父母", "家书与远隔"})
	add(r, "relation", spec{"relation.friend", "朋友知交", "同路之人"})
	add(r, "relation.friend", spec{"relation.friend.old", "故交老友", "久别重逢"})
	add(r, "relation.friend", spec{"relation.friend.new", "新交", "刚刚认识"})
	add(r, "relation.friend", spec{"relation.friend.exiled", "同病相怜", "一起被贬的人"})
	add(r, "relation", spec{"relation.romance", "情爱", "男女之情"})
	add(r, "relation.romance", spec{"relation.romance.longing", "相思", "见不到的人"})
	add(r, "relation.romance", spec{"relation.romance.heartbreak", "伤心", "被辜负"})
	add(r, "relation", spec{"relation.official", "官场往来", "同僚与上级"})
	add(r, "relation.official", spec{"relation.official.patron", "知己上司", "赏识你的人"})
	add(r, "relation.official", spec{"relation.official.rival", "同僚倾轧", "官场的暗流"})
	add(r, "relation", spec{"relation.society", "人群与社会", "在人群中的位置"})
	add(r, "relation.society", spec{"relation.society.crowd", "独处众人间", "热闹里的孤独"})
	add(r, "relation.society", spec{"relation.society.excluded", "被排除在外", "进不去的场"})
	add(r, "relation", spec{"relation.stranger", "与陌生人", "擦肩而过"})

	// ---------------------------------------------------------------- 5. 境遇
	r = model.RootOccasion
	add(r, "", spec{"occasion", "人生境遇与现实处境", "人生的哪个路口"})
	add(r, "occasion", spec{"occasion.exile", "贬谪流放", "被贬到远处"})
	add(r, "occasion.exile", spec{"occasion.exile.hardship", "艰难困顿", "食无肉、病无医"})
	add(r, "occasion.exile", spec{"occasion.exile.grand", "处境渐好", "起复与北归"})
	add(r, "occasion", spec{"occasion.career", "仕途起落", "官场升降"})
	add(r, "occasion.career", spec{"occasion.career.appointed", "受重用", "终于被看见"})
	add(r, "occasion.career", spec{"occasion.career.passed", "被冷落", "不被重用"})
	add(r, "occasion", spec{"occasion.life", "人生关口", "绕不开的那些事"})
	add(r, "occasion.life", spec{"occasion.life.ageing", "渐老", "白发与时光"})
	add(r, "occasion.life", spec{"occasion.life.illness", "病中", "身体的罢工"})
	add(r, "occasion.life", spec{"occasion.life.death", "生死与丧", "至亲的离去"})
	add(r, "occasion.life", spec{"occasion.life.renewal", "重新开始", "再一次出发"})
	add(r, "occasion", spec{"occasion.daily", "日常起居", "普通的一天"})
	add(r, "occasion.daily", spec{"occasion.daily.meal", "饮食", "一碗饭一件事"})
	add(r, "occasion.daily", spec{"occasion.daily.sleep", "睡眠与无眠", "夜里睡不着"})
	add(r, "occasion.daily", spec{"occasion.daily.idle", "无所事事", "一天过去什么都没做"})
	add(r, "occasion", spec{"occasion.feast", "宴饮聚会", "酒桌与欢场"})
	add(r, "occasion.feast", spec{"occasion.feast.drinking", "饮酒", "一杯接一杯"})
	add(r, "occasion.feast", spec{"occasion.feast.music", "听歌赏舞", "笙箫与红袖"})
	add(r, "occasion", spec{"occasion.journey", "远行与归来", "在路上与到站"})
	add(r, "occasion.journey", spec{"occasion.journey.departure", "启程", "背上行囊"})
	add(r, "occasion.journey", spec{"occasion.journey.return", "归乡", "回来了"})
	add(r, "occasion", spec{"occasion.solo", "独自一人", "一个人待着"})

	// ---- 第四层（叶节点）：把更细的日常情境落到可点选的末端
	add(r, "mood.low.tangled", spec{"mood.low.tangled.inexpressible", "说不出口", "明明很在意却开不了口"})
	add(r, "mood.low.tangled", spec{"mood.low.tangled.ambivalent", "又爱又恨", "两种情绪同时成立"})
	add(r, "mood.low.tangled", spec{"mood.low.tangled.ashamed", "羞于面对", "想逃开又不能逃"})
	add(r, "mood.joy.casual", spec{"mood.joy.casual.idle", "无所事事的半天", "偷得浮生半日闲"})
	add(r, "mood.joy.casual", spec{"mood.joy.casual.food", "为一顿好饭高兴", "人间至味"})
	add(r, "mood.joy.release", spec{"mood.joy.release.drunk", "醉后松快", "酒到情绪松开"})
	add(r, "mood.insight.plain", spec{"mood.insight.plain.enough", "知足常乐", "够了，就够了"})
	add(r, "mood.insight.zen", spec{"mood.insight.zen.detached", "袖手旁观", "看而不参与"})

	add(r, "place.nature.river", spec{"place.nature.river.lake", "湖", "水天相接"})
	add(r, "place.nature.river", spec{"place.nature.river.ferry", "渡口", "在水一方"})
	add(r, "place.nature.garden", spec{"place.nature.garden.plum", "梅竹", "案头清供"})
	add(r, "place.city.street", spec{"place.city.street.market", "集市", "人间的热闹"})
	add(r, "place.travel.boat", spec{"place.travel.boat.night", "夜航", "船到江心"})
	add(r, "place.indoor.bed", spec{"place.indoor.bed.insomnia", "辗转难眠", "夜里最诚实的时刻"})

	add(r, "weather.sky.moon", spec{"weather.sky.moon.full", "满月", "千里共婵娟"})
	add(r, "weather.sky.moon", spec{"weather.sky.moon.crescent", "残月", "将圆未圆"})
	add(r, "weather.rain.snow", spec{"weather.rain.snow.first", "初雪", "忽如一夜春风来"})
	add(r, "weather.time.night", spec{"weather.time.night.late", "夜深", "三更四更"})
	add(r, "weather.season.autumn", spec{"weather.season.autumn.chill", "秋凉", "一场秋雨一场寒"})

	add(r, "relation.friend.old", spec{"relation.friend.old.reunion", "久别重逢", "万里归来仍识君"})
	add(r, "relation.friend.old", spec{"relation.friend.old.gone", "故人已远", "唯见长江东流水"})
	add(r, "relation.family.brother", spec{"relation.family.brother.sibling", "与弟", "手足情深"})
	add(r, "relation.romance.longing", spec{"relation.romance.longing.distance", "隔水隔山", "隔断了也要相望"})

	add(r, "occasion.exile.hardship", spec{"occasion.exile.hardship.poverty", "贫病交加", "无肉无药"})
	add(r, "occasion.exile.hardship", spec{"occasion.exile.hardship.solitude", "孤悬海外", "举目皆非故人"})
	add(r, "occasion.life.death", spec{"occasion.life.death.parent", "丧亲之痛", "十年生死两茫茫"})
	add(r, "occasion.life.death", spec{"occasion.life.death.self", "自伤", "此心飘然"})
	add(r, "occasion.life.ageing", spec{"occasion.life.ageing.hair", "白发", "休将白发唱黄鸡"})
	add(r, "occasion.feast.drinking", spec{"occasion.feast.drinking.sober", "醉后独醒", "酒醒人不眠"})

	t.Finalize()
	return t
}
