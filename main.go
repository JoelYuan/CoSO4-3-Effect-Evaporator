package main

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
)

const (
	TARGET_T3 = 60.0  // 出料真实温度必须60.0℃
	TARGET_W3 = 52.30 // 出料目标浓度52.30%
)

// 蒸汽表数据 (绝对压力MPa, 温度℃)
var steamTable = []struct {
	pressure float64 // MPa
	temp     float64 // ℃
}{
	{0.001, 6.69}, {0.002, 17.20}, {0.003, 23.77}, {0.004, 28.66}, {0.005, 32.55},
	{0.006, 35.28}, {0.007, 38.66}, {0.008, 41.16}, {0.009, 43.41}, {0.01, 45.45},
	{0.015, 53.59}, {0.02, 59.66}, {0.025, 64.55}, {0.03, 68.67}, {0.035, 71.75},
	{0.04, 75.41}, {0.045, 78.26}, {0.05, 80.86}, {0.055, 83.24}, {0.060, 85.45},
	{0.065, 87.51}, {0.07, 89.44}, {0.075, 91.26}, {0.08, 92.98}, {0.085, 94.64},
	{0.09, 96.17}, {0.095, 97.66}, {0.10, 98.08}, {0.1013, 100}, {0.105, 101},
	{0.1088, 102}, {0.1127, 103}, {0.1167, 104}, {0.1208, 105}, {0.125, 106},
	{0.1294, 107}, {0.1339, 108}, {0.1385, 109}, {0.1433, 110}, {0.1481, 111},
	{0.1532, 112}, {0.1583, 113}, {0.1636, 114}, {0.1691, 115}, {0.1746, 116},
	{0.1804, 117}, {0.1863, 118}, {0.1923, 119}, {0.1985, 120}, {0.2049, 121},
	{0.2114, 122}, {0.2182, 123}, {0.225, 124}, {0.2321, 125}, {0.2393, 126},
	{0.2467, 127}, {0.2543, 128}, {0.2621, 129}, {0.2701, 130}, {0.2783, 131},
	{0.2867, 132}, {0.2953, 133}, {0.3041, 134}, {0.313, 135}, {0.3222, 136},
	{0.3317, 137}, {0.3414, 138}, {0.3513, 139}, {0.3614, 140}, {0.3718, 141},
	{0.3823, 142}, {0.3931, 143}, {0.4042, 144}, {0.4155, 145}, {0.4271, 146},
	{0.4389, 147}, {0.451, 148}, {0.4633, 149}, {0.476, 150}, {0.4888, 151},
	{0.5021, 152}, {0.5155, 153}, {0.5292, 154}, {0.537, 155}, {0.5577, 156},
	{0.5723, 157}, {0.5872, 158}, {0.6025, 159}, {0.6181, 160}, {0.6339, 161},
	{0.6502, 162}, {0.6666, 163}, {0.6835, 164}, {0.7008, 165}, {0.7183, 166},
	{0.7362, 167}, {0.7544, 168}, {0.773, 169}, {0.792, 170}, {0.8114, 171},
	{0.831, 172}, {0.8511, 173}, {0.8716, 174}, {0.8924, 175}, {0.9137, 176},
	{0.9353, 177}, {0.9573, 178}, {0.9797, 179}, {1.0197, 180}, {1.0259, 181},
	{1.0496, 182}, {1.0737, 183}, {1.0983, 184}, {1.1233, 185}, {1.1487, 186},
	{1.1746, 187}, {1.201, 188}, {1.2278, 189}, {1.2551, 190}, {1.2829, 191},
	{1.3111, 192}, {1.3397, 193}, {1.369, 194}, {1.3987, 195}, {1.4289, 196},
	{1.4596, 197}, {1.4909, 198}, {1.5225, 199}, {1.5548, 200}, {1.5876, 201},
	{1.621, 202}, {1.6548, 203}, {1.6892, 204}, {1.7242, 205}, {1.7597, 206},
	{1.7959, 207}, {1.8326, 208}, {1.8699, 209}, {1.9077, 210}, {1.9462, 211},
	{1.9852, 212}, {2.0248, 213}, {2.065, 214}, {2.1059, 215}, {2.1474, 216},
	{2.1896, 217}, {2.2323, 218}, {2.2757, 219}, {2.3198, 220}, {2.3645, 221},
	{2.4098, 222}, {2.4559, 223}, {2.5026, 224}, {2.55, 225}, {2.5981, 226},
	{2.6469, 227}, {2.6963, 228}, {2.7466, 229}, {2.7975, 230}, {2.8491, 231},
	{2.901, 232}, {2.9546, 233}, {3.0085, 234}, {3.0631, 235}, {3.1185, 236},
	{3.1476, 237}, {3.2316, 238}, {3.2892, 239}, {3.3477, 240}, {3.407, 241},
	{3.467, 242}, {3.5279, 243}, {3.5897, 244}, {3.6522, 245}, {3.7155, 246},
	{3.7797, 247},
}

// 七水合硫酸钴密度表（温度℃ → []{七水质量分数%, 密度g/cm³}）
var densityTable = map[float64][][2]float64{
	20:  {{0, 1.000}, {10, 1.092}, {15, 1.142}, {20, 1.195}, {25, 1.250}, {30, 1.308}, {35, 1.368}, {40, 1.431}, {45, 1.497}, {48, 1.540}, {50, 1.569}, {51, 1.584}, {52, 1.599}},
	40:  {{0, 1.000}, {15, 1.126}, {20, 1.175}, {25, 1.227}, {30, 1.282}, {35, 1.340}, {40, 1.401}, {45, 1.465}, {48, 1.505}, {50, 1.533}, {51, 1.547}, {52, 1.561}},
	50:  {{0, 1.000}, {20, 1.160}, {25, 1.210}, {30, 1.263}, {35, 1.319}, {40, 1.378}, {45, 1.440}, {48, 1.478}, {50, 1.505}, {51, 1.519}, {52, 1.533}},
	55:  {{0, 1.000}, {30, 1.247}, {34, 1.293}, {38, 1.345}, {42, 1.400}, {46, 1.458}, {49, 1.500}, {50, 1.515}, {51, 1.530}, {51.8, 1.540}},
	60:  {{0, 1.000}, {32, 1.268}, {36, 1.316}, {40, 1.368}, {44, 1.423}, {48, 1.482}, {50, 1.512}, {51, 1.527}, {52, 1.542}, {53, 1.557}},
	80:  {{0, 0.992}, {40, 1.315}, {45, 1.367}, {48, 1.405}, {50, 1.433}, {51, 1.447}, {52, 1.461}},
	100: {{0, 0.980}, {45, 1.330}, {48, 1.365}, {50, 1.392}, {51, 1.405}, {52, 1.418}},
}

func BPR(w float64) float64 {
	return 0.00028*w*w*w - 0.021*w*w + 0.78*w - 8.1
}

// 根据温度查找压力 (线性插值)
func tempToPressure(temp float64) float64 {
	if temp < steamTable[0].temp {
		return steamTable[0].pressure
	}
	if temp > steamTable[len(steamTable)-1].temp {
		return steamTable[len(steamTable)-1].pressure
	}

	// 找到相邻的两个点
	for i := 0; i < len(steamTable)-1; i++ {
		if temp >= steamTable[i].temp && temp <= steamTable[i+1].temp {
			// 线性插值
			ratio := (temp - steamTable[i].temp) / (steamTable[i+1].temp - steamTable[i].temp)
			return steamTable[i].pressure + ratio*(steamTable[i+1].pressure-steamTable[i].pressure)
		}
	}
	return steamTable[0].pressure
}

// 根据压力查找温度 (线性插值)
func pressureToTemp(pressure float64) float64 {
	if pressure < steamTable[0].pressure {
		return steamTable[0].temp
	}
	if pressure > steamTable[len(steamTable)-1].pressure {
		return steamTable[len(steamTable)-1].temp
	}

	// 找到相邻的两个点
	for i := 0; i < len(steamTable)-1; i++ {
		if pressure >= steamTable[i].pressure && pressure <= steamTable[i+1].pressure {
			// 线性插值
			ratio := (pressure - steamTable[i].pressure) / (steamTable[i+1].pressure - steamTable[i].pressure)
			return steamTable[i].temp + ratio*(steamTable[i+1].temp-steamTable[i].temp)
		}
	}
	return steamTable[0].temp
}

// 根据温度和浓度查找密度 (双线性插值)
func densityFromTempAndConc(temp, conc float64) float64 {
	// 获取温度附近的两个温度点
	tempKeys := make([]float64, 0, len(densityTable))
	for t := range densityTable {
		tempKeys = append(tempKeys, t)
	}
	sort.Float64s(tempKeys)

	// 找到相邻的两个温度点
	var lowerTemp, upperTemp float64
	lowerFound := false
	upperFound := false

	for _, t := range tempKeys {
		if t <= temp {
			lowerTemp = t
			lowerFound = true
		}
		if t >= temp && !upperFound {
			upperTemp = t
			upperFound = true
			break
		}
	}

	if !lowerFound && !upperFound {
		return 1.0 // 默认密度
	}

	// 如果温度正好匹配其中一个点
	if lowerTemp == upperTemp {
		return densityAtTemp(temp, conc)
	}

	// 双线性插值
	densityLower := densityAtTemp(lowerTemp, conc)
	densityUpper := densityAtTemp(upperTemp, conc)

	// 温度方向的插值
	if lowerTemp == upperTemp {
		return densityLower
	}

	ratio := (temp - lowerTemp) / (upperTemp - lowerTemp)
	return densityLower + ratio*(densityUpper-densityLower)
}

// 在特定温度下根据浓度查找密度 (线性插值)
func densityAtTemp(temp, conc float64) float64 {
	data, exists := densityTable[temp]
	if !exists {
		// 如果温度点不存在，使用最近的温度点
		tempKeys := make([]float64, 0, len(densityTable))
		for t := range densityTable {
			tempKeys = append(tempKeys, t)
		}
		sort.Float64s(tempKeys)

		closestTemp := tempKeys[0]
		minDiff := math.Abs(temp - tempKeys[0])
		for _, t := range tempKeys {
			if math.Abs(temp-t) < minDiff {
				minDiff = math.Abs(temp - t)
				closestTemp = t
			}
		}
		data = densityTable[closestTemp]
	}

	// 找到浓度附近的两个点
	var lowerConc, upperConc, lowerDensity, upperDensity float64
	foundLower := false
	foundUpper := false

	for _, entry := range data {
		c, d := entry[0], entry[1]
		if c <= conc {
			lowerConc, lowerDensity = c, d
			foundLower = true
		}
		if c >= conc && !foundUpper {
			upperConc, upperDensity = c, d
			foundUpper = true
			break
		}
	}

	if !foundLower && !foundUpper {
		return 1.0 // 默认密度
	}

	// 如果浓度正好匹配其中一个点
	if lowerConc == upperConc {
		return lowerDensity
	}

	// 浓度方向的线性插值
	ratio := (conc - lowerConc) / (upperConc - lowerConc)
	return lowerDensity + ratio*(upperDensity-lowerDensity)
}

// 根据温度和密度查找浓度 (线性插值)
func concFromTempAndDensity(temp, density float64) float64 {
	data, exists := densityTable[temp]
	if !exists {
		// 如果温度点不存在，使用最近的温度点
		tempKeys := make([]float64, 0, len(densityTable))
		for t := range densityTable {
			tempKeys = append(tempKeys, t)
		}
		sort.Float64s(tempKeys)

		closestTemp := tempKeys[0]
		minDiff := math.Abs(temp - tempKeys[0])
		for _, t := range tempKeys {
			if math.Abs(temp-t) < minDiff {
				minDiff = math.Abs(temp - t)
				closestTemp = t
			}
		}
		data = densityTable[closestTemp]
	}

	// 找到密度附近的两个点
	var lowerConc, upperConc, lowerDensity, upperDensity float64
	foundLower := false
	foundUpper := false

	for _, entry := range data {
		c, d := entry[0], entry[1]
		if d <= density {
			lowerConc, lowerDensity = c, d
			foundLower = true
		}
		if d >= density && !foundUpper {
			upperConc, upperDensity = c, d
			foundUpper = true
			break
		}
	}

	if !foundLower && !foundUpper {
		return 0.0 // 默认浓度
	}

	// 如果密度正好匹配其中一个点
	if lowerDensity == upperDensity {
		return lowerConc
	}

	// 密度方向的线性插值
	ratio := (density - lowerDensity) / (upperDensity - lowerDensity)
	return lowerConc + ratio*(upperConc-lowerConc)
}

type EvaporatorResult struct {
	F0, F1, F2, F3                         float64 // 各效流量
	W0, W1, W2, W3                         float64 // 各效浓度
	V1, V2, V3                             float64 // 各效蒸发量
	BPR1, BPR2, BPR3                       float64 // 各效沸点升高
	Density0, Density1, Density2, Density3 float64 // 各效密度
	Pset                                   float64 // 真空设定值
	T3                                     float64 // 第三效温度
	R1, R2, R3                             float64 // 各效负荷比例
}

func calculateEvaporator(F0, w0_percent, TS, T1_out, T2_out, level3 float64) EvaporatorResult {
	// 计算出料流量
	F3 := F0 * w0_percent / TARGET_W3
	Vtotal := F0 - F3

	// 基于温差动态计算负荷分配比例
	// 温差越大，负荷越大
	tempDiff1 := TS - T1_out        // 第一效温差
	tempDiff2 := T1_out - T2_out    // 第二效温差
	tempDiff3 := T2_out - TARGET_T3 // 第三效温差 (修正：使用目标温度60℃)

	// 避免负温差
	if tempDiff1 <= 0 {
		tempDiff1 = 1
	}
	if tempDiff2 <= 0 {
		tempDiff2 = 1
	}
	if tempDiff3 <= 0 {
		tempDiff3 = 1
	}

	// 负荷比例与温差成正比
	totalDiff := tempDiff1 + tempDiff2 + tempDiff3
	r1 := tempDiff1 / totalDiff
	r2 := tempDiff2 / totalDiff
	r3 := tempDiff3 / totalDiff

	// 按比例分配蒸发量
	V1 := Vtotal * r1
	V2 := Vtotal * r2
	V3 := Vtotal * r3

	// 计算各效流量和浓度
	F1 := F0 - V1
	w1 := w0_percent * F0 / F1
	F2 := F1 - V2
	w2 := w1 * F1 / F2
	F3calc := F2 - V3
	w3 := w2 * F2 / F3calc

	// 计算沸点升高 - 修正：第三效沸点升高应基于目标浓度
	bpr1 := BPR(w1)
	bpr2 := BPR(w2)
	bpr3 := BPR(TARGET_W3) // 第三效沸点升高基于目标浓度52.30%

	// 计算各效密度（假设温度为60℃）
	density0 := densityFromTempAndConc(60, w0_percent)
	density1 := densityFromTempAndConc(60, w1)
	density2 := densityFromTempAndConc(60, w2)
	density3 := densityFromTempAndConc(60, w3)

	static := level3 * 0.35
	targetTemp := TARGET_T3 - bpr3 - static // 这里才减去沸点升高和静压头

	// 使用精确蒸汽表查找压力
	Pset := tempToPressure(targetTemp)

	return EvaporatorResult{
		F0: F0, F1: F1, F2: F2, F3: F3calc,
		W0: w0_percent, W1: w1, W2: w2, W3: w3,
		V1: V1, V2: V2, V3: V3,
		BPR1: bpr1, BPR2: bpr2, BPR3: bpr3,
		Density0: density0, Density1: density1, Density2: density2, Density3: density3,
		Pset: Pset * 1000, // 转换为 kPa
		T3:   targetTemp,
		R1:   r1, R2: r2, R3: r3,
	}
}

func readFloat(prompt string) float64 {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print(prompt)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		if input == "" {
			fmt.Println("   ↑ 不能为空，请重新输入")
			continue
		}
		val, err := strconv.ParseFloat(input, 64)
		if err != nil {
			fmt.Println("   ↑ 输入无效，请输入数字")
			continue
		}
		return val
	}
}

func main() {
	fmt.Println("==================================================")
	fmt.Println("      七水硫酸钴三效蒸发排班计算器")
	fmt.Println(" （动态负荷分配版 - 基于介质温度）")
	fmt.Println("==================================================\n")

	// 1. 进料流量
	F0 := readFloat("1. 进料流量 (kg/h)        → ")

	// 2. 进料状态（浓度优先，密度备选）
	fmt.Println("\n2. 进料浓度方式请选择：")
	fmt.Println("   [1] 直接输入浓度 %")
	fmt.Println("   [2] 输入温度+密度自动换算")
	choice := readFloat("   请选择 1 或 2 → ")

	var w0_percent float64
	if choice == 1 {
		w0_percent = readFloat("   进料浓度 (%)            → ")
	} else {
		tIn := readFloat("   进料温度 (℃)            → ")
		dIn := readFloat("   进料密度 (g/cm³)        → ")
		// 使用密度表进行双向插值计算浓度
		w0_percent = concFromTempAndDensity(tIn, dIn)
		fmt.Printf("   → 通过密度表插值得到浓度 ≈ %.2f%%\n", w0_percent)
	}

	// 3. 加热蒸汽温度（第一效）
	TS := readFloat("\n3. 第一效加热蒸汽温度 (℃) → ")

	// 4. 第一效出液温度（用于二次汽）
	T1_out := readFloat("   第一效出液温度 (℃)       → ")

	// 5. 第二效出液温度（用于二次汽）
	T2_out := readFloat("   第二效出液温度 (℃)       → ")

	// 6. 末效液位（静压头）
	level3 := readFloat("\n6. 当前第三效液位 (m)     → ")

	fmt.Println("\n正在计算……\n")
	// ====================== 开始计算 ======================
	result := calculateEvaporator(F0, w0_percent, TS, T1_out, T2_out, level3)

	// ====================== 输出完整排班表 ======================
	fmt.Printf("╔══════════════════════════════════════════════════════════════╗\n")
	fmt.Printf("║                   今日硫酸钴三效排班计算结果                   ║\n")
	fmt.Printf("║                   （液温参评估的负荷分配版）                   ║\n")
	fmt.Printf("╚══════════════════════════════════════════════════════════════╝\n\n")

	fmt.Printf("进料            %.0f kg/h， 浓度 %.2f%%， 蒸汽温度 %.1f℃\n", F0, w0_percent, TS)
	fmt.Printf("总蒸发水量      %.0f kg/h → 出料 %.0f kg/h（52.30%%）\n\n", result.V1+result.V2+result.V3, result.F3)

	fmt.Printf("负荷分配        %.2f : %.2f : %.2f\n", result.R1*100, result.R2*100, result.R3*100)
	fmt.Printf("第一效蒸发      %.0f kg/h\n", result.V1)
	fmt.Printf("第二效蒸发      %.0f kg/h\n", result.V2)
	fmt.Printf("第三效蒸发      %.0f kg/h\n\n", result.V3)

	fmt.Printf("各效浓度一览（发给操作工）\n")
	fmt.Printf("┌───────────────────────────────────────────────────────┐\n")
	fmt.Printf("│ 效号 │ 流量(kg/h) │ 浓度(%%) │ 密度(g/cm³) │ 沸点升高(℃) │\n")
	fmt.Printf("├───────────────────────────────────────────────────────┤\n")
	fmt.Printf("│ 进料   │ %8.0f  │ %6.2f │ %9.3f │     -        │\n", result.F0, result.W0, result.Density0)
	fmt.Printf("│ 第一效 │ %8.0f  │ %6.2f │ %9.3f │ %8.2f      │\n", result.F1, result.W1, result.Density1, result.BPR1)
	fmt.Printf("│ 第二效 │ %8.0f  │ %6.2f │ %9.3f │ %8.2f      │\n", result.F2, result.W2, result.Density2, result.BPR2)
	fmt.Printf("│ 出料   │ %8.0f  │ %6.2f │ %9.3f │ %8.2f      │\n", result.F3, result.W3, result.Density3, result.BPR3)
	fmt.Printf("└─────────────────────────────────────────────────────┘\n\n")

	fmt.Printf("末效真空设定值（请下发DCS！！！）\n")
	fmt.Printf("────────────────────────────────────────────────────────\n")
	fmt.Printf("第三效目标温度           %.2f ℃\n", result.T3)
	fmt.Printf("第三效出料浓度           %.2f %%\n", result.W3) // 显示实际出料浓度
	fmt.Printf("沸点升高 BPR             %.2f ℃\n", result.BPR3)
	fmt.Printf("静压头修正（液位%.2fm）   %.2f ℃\n", level3, level3*0.35)
	fmt.Printf("\n")
	fmt.Printf("气相绝压设定值           %.3f kPa\n", result.Pset)
	fmt.Printf("真空表示数               %.0f mmHg\n", 760-result.Pset*7.5006/1000)
	fmt.Printf("真空度（表压）           %.1f kPa\n", 101.325-result.Pset/1000)
	fmt.Printf("────────────────────────────────────────────────────────\n")
	fmt.Printf("今日末效真空必须设定 %.3f kPa（绝压），才能保证60.0℃！\n\n", result.Pset)

	fmt.Printf("按回车键退出...")
	fmt.Scanln()
}
