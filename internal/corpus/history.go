package corpus

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/sumate/sumate/internal/model"
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
func stageByPlace(place string, era model.Era) model.LifeStage {
	switch place {
	case "湖北黄冈":
		return model.StageHuangzhou
	case "广东惠州", "海南儋州":
		if place == "广东惠州" {
			return model.StageHuizhou
		}
		return model.StageDanzhou
	case "江苏常州":
		return model.StageLateYears
	}
	switch era {
	case "元丰":
		return model.StageReformConflict
	case "元祐":
		return model.StageReturn
	case "绍圣", "元符":
		return model.StageLateYears
	case "熙宁":
		return model.StageInOffice
	case "治平":
		return model.StageEarlyCareer
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

// DeriveHistory infers the historical background of a poem from its 词题.
func DeriveHistory(p model.Poem) model.Background {
	bg := model.Background{Source: p.Title}

	era, base, ok := findEra(p.Title)
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

	// year ordinal
	year := base
	yearText := string(era)
	if m := ordinalRe.FindStringSubmatch(p.Title); m != nil {
		if n := chineseNum(m[1]); n > 0 {
			year = base + n - 1
			yearText = string(era) + m[1] + "年"
		}
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
	if bg.Year != "" {
		b.WriteString("作于" + bg.Year + "年")
	} else if bg.Era != "" {
		b.WriteString("作于" + string(bg.Era) + "年间")
	}
	if bg.Place != "" {
		b.WriteString("，" + bg.Place)
	}
	b.WriteString("。")
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
