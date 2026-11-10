package corpus

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/incurl/sumate/internal/model"
)

// Era markers in the 词题/小序, mapped to the reign era and its first year.
var eraMarkers = []struct {
	era  model.Era
	year int
	tags []string
}{
	{"治平", 1064, nil},
	{"熙宁", 1068, nil},
	{"元丰", 1078, nil},
	{"元祐", 1086, nil},
	{"绍圣", 1094, nil},
	{"元符", 1098, nil},
	{"建中靖国", 1101, nil},
	{"崇宁", 1102, nil},
}

// wellKnown holds the historical facts of a famous poem.
type wellKnown struct {
	era   model.Era
	year  int
	place string
}

// wellKnownPoems records the historical facts of the famous poems.
//
// The 词题 of the 四库 edition often omits what a modern reader expects:
// 《定风波》 says only "三月七日沙湖道中遇雨", but the poem was written on
// the road to 沙湖 in 黄州 in 1082.  The table is keyed by the 词牌 together
// with a distinctive phrase from the verse, so that the match is exact.
var wellKnownPoems = map[string]wellKnown{
	"水调歌头|明月几时有":    {era: "熙宁", year: 1076, place: "山东密州"},
	"定风波|莫听穿林打叶声":   {era: "元丰", year: 1082, place: "湖北黄冈"},
	"念奴娇|大江东去浪淘尽":   {era: "元丰", year: 1082, place: "湖北黄冈"},
	"江城子|十年生死两茫茫":   {era: "熙宁", year: 1075, place: "山东密州"},
	"浣溪沙|谁道人生无再少":   {era: "元丰", year: 1082, place: "湖北黄冈"},
	"卜算子|拣尽寒枝不肯栖":   {era: "绍圣", year: 1095, place: "广东惠州"},
	"临江仙|小舟从此逝":     {era: "元丰", year: 1083, place: "湖北黄冈"},
	"浣溪沙|此心安处是吾乡":   {era: "元符", year: 1099, place: "广东惠州"},
	"西江月|世事一场大梦":    {era: "元丰", year: 1083, place: "湖北黄冈"},
	"鹧鸪天|林断山明竹隐墙":   {era: "元丰", year: 1080, place: "山东密州"},
	"水龙吟|似花还似非花":    {era: "元丰", year: 1083, place: "湖北黄冈"},
	"贺新郎|乳燕飞华屋":     {era: "元祐", year: 1090, place: "浙江杭州"},
	"行香子|述怀":        {era: "元丰", year: 1082, place: "湖北黄冈"},
	"永遇乐|长忆别时景":     {era: "熙宁", year: 1077, place: "江苏徐州"},
	"满庭芳|蜗角虚名蝇头微利":  {era: "元祐", year: 1087, place: "山东密州"},
	"八声甘州|有情风万里卷潮来": {era: "元祐", year: 1087, place: "江苏徐州"},
	"定风波|竹杖芒鞋轻胜马":   {era: "元丰", year: 1082, place: "湖北黄冈"},
	"蝶恋花|花褪残红青杏了":   {era: "元丰", year: 1084, place: "湖北黄冈"},
	"虞美人|春花秋月何时了":   {era: "熙宁", year: 1076, place: "山东密州"},
	"临江仙|夜饮东坡醒复醉":   {era: "元丰", year: 1083, place: "湖北黄冈"},
	"行香子|七里滩":       {era: "元丰", year: 1082, place: "浙江杭州"},
	"念奴娇|我醉拍手狂歌":    {era: "元祐", year: 1083, place: "浙江杭州"},
	"水调歌头|落红不是无情物":  {era: "元丰", year: 1083, place: "湖北黄冈"},
	"西江月|顷在黄州":      {era: "元丰", year: 1083, place: "湖北黄冈"},
	"卜算子|缺月挂疏桐":     {era: "元丰", year: 1082, place: "湖北黄冈"},
	"江城子|老夫聊发少年狂":   {era: "熙宁", year: 1075, place: "山东密州"},
	"望江南|春未老":       {era: "元丰", year: 1082, place: "湖北黄冈"},
	"定风波|莫听穿林":      {era: "元丰", year: 1082, place: "湖北黄冈"},
	"江神子|江神子":       {era: "熙宁", year: 1075, place: "山东密州"},
	"满江红|江汉西来":      {era: "元丰", year: 1080, place: "江苏徐州"},
	"念奴娇|凭高眺远见长空":   {era: "元祐", year: 1083, place: "浙江杭州"},
	"醉翁操|":          {},
}

// Place markers, mapped to a modern place description.
var placeMarkers = []struct {
	marker string
	place  string
}{
	{"黄州", "湖北黄冈"},
	{"惠州", "广东惠州"},
	{"儋州", "海南儋州"},
	{"徐州", "江苏徐州"},
	{"密州", "山东诸城"},
	{"杭州", "浙江杭州"},
	{"湖州", "浙江湖州"},
	{"颍州", "安徽阜阳"},
	{"常州", "江苏常州"},
	{"苏州", "江苏苏州"},
	{"泗州", "江苏盱眙"},
	{"京口", "江苏镇江"},
	{"润州", "江苏镇江"},
	{"真州", "江苏仪征"},
	{"高邮", "江苏高邮"},
	{"扬州", "江苏扬州"},
	{"泰州", "江苏泰州"},
	{"汝州", "河南汝州"},
	{"登州", "山东蓬莱"},
	{"儋耳", "海南儋州"},
	{"岭南", "岭南一带"},
	{"虔州", "江西赣州"},
	{"永州", "湖南永州"},
	{"潭州", "湖南长沙"},
	{"苏州", "江苏苏州"},
}

// eraRe finds a reign era plus an optional ordinal year.
var eraRe = regexp.MustCompile(`(元丰|元祐|熙宁|绍圣|元符|治平|崇宁|建中靖国|嘉祐|乾道|淳熙|庆历|宝元)`)

// yearRe finds `某年` / `某月某日` ordinals near an era.
var ordinalRe = regexp.MustCompile(`([元一二三四五六七八九十]+|\d+)\s*年`)

// dayRe finds `某月某日`.
var dayRe = regexp.MustCompile(`([一二三四五六七八九十]+|\d+)\s*月\s*([一二三四五六七八九十]+|\d+)\s*日`)

// eraRe finds the era in a title.
func findEra(title string) (model.Era, int, bool) {
	for _, m := range eraMarkers {
		if strings.Contains(title, string(m.era)) {
			return m.era, m.year, true
		}
	}
	return "", 0, false
}

// findPlace finds the place of composition in a title.
func findPlace(title string) string {
	for _, p := range placeMarkers {
		if strings.Contains(title, p.marker) {
			return p.place
		}
	}
	return ""
}

// stageByPlace infers the life stage from the place of composition.
//
// The place is the stronger signal: Su Dongpo's 黄州 years are the only ones
// in which he lived at 黄冈, no matter which era the poem belongs to.
func stageByPlace(place string, era model.Era) model.LifeStage {
	switch place {
	case "湖北黄冈":
		return model.StageHuangzhou
	case "广东惠州":
		return model.StageHuizhou
	case "海南儋州":
		return model.StageDanzhou
	case "江苏常州":
		return model.StageLateYears
	case "山东密州":
		return model.StageInOffice
	case "江苏徐州":
		return model.StageInOffice
	case "浙江杭州":
		return model.StageInOffice
	}
	switch era {
	case "治平":
		return model.StageEarlyCareer
	case "熙宁":
		return model.StageInOffice
	case "元丰":
		return model.StageReformConflict
	case "元祐":
		return model.StageReturn
	case "绍圣":
		return model.StageHuizhou
	case "元符":
		return model.StageDanzhou
	}
	return model.StageLeisure
}

// chineseNum converts a Chinese numeral to an int.
func chineseNum(s string) int {
	if n, err := strconv.Atoi(s); err == nil {
		return n
	}
	r := []rune(s)
	digit := func(c rune) int {
		switch c {
		case '零':
			return 0
		case '一':
			return 1
		case '二':
			return 2
		case '三':
			return 3
		case '四':
			return 4
		case '五':
			return 5
		case '六':
			return 6
		case '七':
			return 7
		case '八':
			return 8
		case '九':
			return 9
		}
		return 0
	}
	switch len(r) {
	case 0:
		return 0
	case 1:
		return digit(r[0])
	case 2:
		if r[0] == '十' {
			return 10
		}
		if r[1] == '十' {
			return digit(r[0]) * 10
		}
		return digit(r[0]) + digit(r[1])
	case 3:
		// e.g. 三十五
		if r[1] == '十' {
			return digit(r[0])*10 + digit(r[2])
		}
	}
	return 0
}

// The Chinese sexagenary cycle, used to resolve the cyclical year names that
// appear in 词题 ("丙辰中秋") into a reign era and a western year.
//
// The cycle repeats every 60 years; 4 CE was 甲子, so the index of a year is
// (year - 4) mod 60.
var heavenlyStems = []string{"甲", "乙", "丙", "丁", "戊", "己", "庚", "辛", "壬", "癸"}
var earthlyBranches = []string{"子", "丑", "寅", "卯", "辰", "巳", "午", "未", "申", "酉", "戌", "亥"}

// cyclicalYears maps a cyclical year name to the (era, year) pairs it can
// refer to inside Su Dongpo's lifetime.  The cycle repeats every 60 years,
// so a name alone is ambiguous; the era disambiguates it.
//
// The table is derived from the cycle itself -- see cyclicalCandidates.
var cyclicalYears = map[string][]model.EraYear{}

func init() {
	// Su Dongpo lived from 1037 to 1101; only the years inside his
	// lifetime are relevant.
	for y := 1037; y <= 1101; y++ {
		name := ganzhiOf(y)
		cyclicalYears[name] = append(cyclicalYears[name],
			model.EraYear{Era: eraOfYear(y), Year: y})
	}
}

// ganzhiOf returns the cyclical name of a western year.
func ganzhiOf(year int) string {
	idx := (year - 4) % 60
	if idx < 0 {
		idx += 60
	}
	return heavenlyStems[idx%10] + earthlyBranches[idx%12]
}

// cyclicalCandidates returns the (era, year) pairs a cyclical name can refer
// to inside Su Dongpo's lifetime.
func cyclicalCandidates(name string) []model.EraYear {
	if list, ok := cyclicalYears[name]; ok {
		return list
	}
	// Fall back to the cycle itself.
	var out []model.EraYear
	for y := 1007; y <= 1101; y++ {
		if ganzhiOf(y) == name {
			out = append(out, model.EraYear{Era: eraOfYear(y), Year: y})
		}
	}
	return out
}

// eraOfYear returns the reign era a western year falls in.
func eraOfYear(year int) model.Era {
	switch {
	case year <= 1067:
		return "治平"
	case year <= 1077:
		return "熙宁"
	case year <= 1085:
		return "元丰"
	case year <= 1093:
		return "元祐"
	case year <= 1097:
		return "绍圣"
	case year <= 1100:
		return "元符"
	}
	return "建中靖国"
}

// DeriveHistory infers the historical background of a poem from its 词题.
func DeriveHistory(p model.Poem) model.Background {
	bg := model.Background{Source: p.Title}

	wellKnownPlace := ""
	fromTable := false
	era, base, ok := findEra(p.Title)
	if !ok {
		// A few famous poems carry their date in the scholarly tradition
		// rather than in the 词题.
		if wk, found := lookupWellKnown(p); found {
			era, base = wk.era, wk.year
			wellKnownPlace = wk.place
			fromTable = true
			ok = true
		}
	}
	if !ok {
		// A 词题 may name only the cyclical year ("丙辰中秋"); the era and
		// the western year can be recovered from the sexagenary cycle.
		if name, cyc, found := findCyclicalYear(p.Title); found {
			bg.Year = name + "年"
			bg.Era = cyc.Era
			bg.GregorianYear = cyc.Year
			bg.Place = findPlace(p.Title)
			if bg.Place == "" {
				bg.Place = wellKnownPlace
			}
			bg.Stage = stageByPlace(bg.Place, cyc.Era)
			bg.Historical = buildHistorical(bg)
			return bg
		}
	}
	if !ok {
		// no explicit era: infer from the place
		place := findPlace(p.Title)
		if place != "" {
			bg.Place = place
			bg.Stage = stageByPlace(place, "")
			bg.Historical = "作于" + place + "，具体年月未见于词题。"
		} else {
			bg.Stage = model.StageLeisure
			bg.Historical = "词题未纪年，按苏轼晚年作品系之。"
		}
		return bg
	}

	bg.Era = era
	bg.Place = findPlace(p.Title)
	if bg.Place == "" {
		bg.Place = wellKnownPlace
	}

	// The year label.  When the 词题 gives an ordinal ("元丰五年") it is used
	// verbatim; otherwise the era itself names the year.
	year := base
	yearText := string(era) + "年间"
	if m := ordinalRe.FindStringSubmatch(p.Title); m != nil {
		if n := chineseNum(m[1]); n > 0 {
			year = base + n - 1
			yearText = string(era) + m[1] + "年"
		}
	} else if fromTable {
		yearText = string(era) + "年间"
	}
	bg.Year = yearText
	bg.GregorianYear = year
	bg.Stage = stageByPlace(bg.Place, era)

	// a day-level date makes the note more precise
	if m := dayRe.FindStringSubmatch(p.Title); m != nil {
		bg.Historical = bg.Historical + "（" + m[1] + "月" + m[2] + "日）"
	}

	bg.Historical = buildHistorical(bg)
	return bg
}

// buildHistorical composes a prose description of the background.
func buildHistorical(bg model.Background) string {
	var b strings.Builder
	switch {
	case bg.Year != "" && strings.HasSuffix(bg.Year, "年"):
		// The label already carries the era ("元丰五年").
		b.WriteString("作于" + bg.Year)
	case bg.Year != "":
		b.WriteString("作于" + bg.Year)
	case bg.Era != "":
		b.WriteString("作于" + string(bg.Era) + "年间")
	}
	if bg.Place != "" {
		b.WriteString("，" + bg.Place)
	}
	b.WriteString("。")
	if bg.GregorianYear > 0 {
		b.WriteString(fmt.Sprintf("（公元%d年）", bg.GregorianYear))
	}
	switch bg.Stage {
	case model.StageHuangzhou:
		b.WriteString("时苏轼谪居黄州，境况清贫而心境渐趋旷达。")
	case model.StageHuizhou:
		b.WriteString("时苏轼贬谪惠州，处境更为困顿而胸次愈发宽厚。")
	case model.StageDanzhou:
		b.WriteString("时苏轼渡海儋州，已近暮年，文字愈见平淡天真。")
	case model.StageReturn:
		b.WriteString("时苏轼起复京师，仕途重开而旧疾未愈。")
	case model.StageReformConflict:
		b.WriteString("时值新旧党争，苏轼外放多年，词风渐由绮丽转向豪放。")
	case model.StageInOffice:
		b.WriteString("时苏轼方值仕途，境遇顺遂，笔意清朗。")
	case model.StageLateYears:
		b.WriteString("时苏轼晚年北归途中，阅尽人事，语气归于平淡。")
	case model.StageTravel:
		b.WriteString("苏轼行旅途中所作。")
	default:
		b.WriteString("苏轼退居闲适之时，作品多寓真趣于日常。")
	}
	return b.String()
}

// findCyclicalYear finds a cyclical year name in a title.
//
// The name must not be part of a longer word, so the characters around it are
// checked.
func findCyclicalYear(title string) (string, model.EraYear, bool) {
	// Iterate over the candidates in a fixed order so that the result never
	// depends on Go's map iteration order.
	names := make([]string, 0, len(cyclicalYears))
	for name := range cyclicalYears {
		names = append(names, name)
	}
	sort.Strings(names)

	place := findPlace(title)
	for _, name := range names {
		list := cyclicalYears[name]
		if len(list) == 0 || !strings.Contains(title, name) {
			continue
		}
		// Prefer the candidate whose era agrees with the place named in the
		// title; otherwise take the most recent one.
		for _, c := range list {
			if place != "" && placeConsistent(place, c) {
				return name, c, true
			}
		}
		return name, list[len(list)-1], true
	}
	return "", model.EraYear{}, false
}

// placeConsistent reports whether a place could have been visited in the year
// of an era.
func placeConsistent(place string, c model.EraYear) bool {
	switch place {
	case "湖北黄冈":
		return c.Era == "元丰"
	case "广东惠州":
		return c.Era == "绍圣"
	case "海南儋州":
		return c.Era == "元符"
	case "江苏常州":
		return c.Year >= 1101
	}
	return false
}

// lookupWellKnown finds the historical facts of a famous poem.
//
// The key is the 词牌 plus a distinctive phrase from the verse; the 四库
// 词题 alone is often too terse to identify the poem.
func lookupWellKnown(p model.Poem) (wellKnown, bool) {
	// The longest matching phrase wins, so that a specific key such as
	// 「卜算子|拣尽寒枝不肯栖」 is preferred over a shorter one that is a
	// prefix of it.  Go's map iteration order is random, so the comparison
	// has to be explicit.
	bestKey := ""
	var best wellKnown
	for key, wk := range wellKnownPoems {
		tune, phrase, _ := strings.Cut(key, "|")
		if tune != p.Tune {
			continue
		}
		if phrase != "" && !strings.Contains(p.Text, phrase) {
			continue
		}
		if len(phrase) > len(bestKey) {
			bestKey, best = phrase, wk
		}
	}
	if bestKey == "" {
		return wellKnown{}, false
	}
	return best, true
}
