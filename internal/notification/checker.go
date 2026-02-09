package notification

import "github.com/example/LottoSmash/internal/lotto"

// CheckWinning 추천번호와 당첨번호를 비교하여 당첨 결과를 반환
// 한국 로또 6/45 등수 기준:
//   - 1등: 6개 번호 일치
//   - 2등: 5개 번호 일치 + 보너스 번호 일치
//   - 3등: 5개 번호 일치
//   - 4등: 4개 번호 일치
//   - 5등: 3개 번호 일치
func CheckWinning(draw *lotto.LottoDraw, rec RecommendationRow) WinningCheck {
	winNums := map[int]bool{
		draw.Num1: true,
		draw.Num2: true,
		draw.Num3: true,
		draw.Num4: true,
		draw.Num5: true,
		draw.Num6: true,
	}

	var matched []int
	for _, n := range rec.Numbers {
		if winNums[n] {
			matched = append(matched, n)
		}
	}

	bonusMatched := false
	for _, n := range rec.Numbers {
		if n == draw.BonusNum && !winNums[n] {
			bonusMatched = true
			break
		}
	}

	matchCount := len(matched)
	var prizeRank *int

	switch {
	case matchCount == 6:
		rank := 1
		prizeRank = &rank
	case matchCount == 5 && bonusMatched:
		rank := 2
		prizeRank = &rank
	case matchCount == 5:
		rank := 3
		prizeRank = &rank
	case matchCount == 4:
		rank := 4
		prizeRank = &rank
	case matchCount == 3:
		rank := 5
		prizeRank = &rank
	}

	return WinningCheck{
		RecommendationID: rec.ID,
		UserID:           rec.UserID,
		DrawNo:           draw.DrawNo,
		MatchedNumbers:   matched,
		MatchedCount:     matchCount,
		BonusMatched:     bonusMatched,
		PrizeRank:        prizeRank,
	}
}

// PrizeRankName 등수를 한글 이름으로 반환
func PrizeRankName(rank int) string {
	switch rank {
	case 1:
		return "1등"
	case 2:
		return "2등"
	case 3:
		return "3등"
	case 4:
		return "4등"
	case 5:
		return "5등"
	default:
		return "미당첨"
	}
}
