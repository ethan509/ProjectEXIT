-- 분석 방법 메타데이터 테이블
CREATE TABLE IF NOT EXISTS analysis_methods (
    id              SERIAL PRIMARY KEY,
    code            VARCHAR(30) NOT NULL UNIQUE,
    name            VARCHAR(50) NOT NULL,
    description     TEXT,
    category        VARCHAR(20) NOT NULL,
    is_active       BOOLEAN DEFAULT true,
    sort_order      INTEGER DEFAULT 0,
    created_at      TIMESTAMP DEFAULT NOW(),
    updated_at      TIMESTAMP DEFAULT NOW()
);

-- 인덱스
CREATE INDEX IF NOT EXISTS idx_analysis_methods_code ON analysis_methods(code);
CREATE INDEX IF NOT EXISTS idx_analysis_methods_active ON analysis_methods(is_active);
CREATE INDEX IF NOT EXISTS idx_analysis_methods_sort ON analysis_methods(sort_order);

-- 초기 데이터 삽입
INSERT INTO analysis_methods (code, name, description, category, sort_order) VALUES
('NUMBER_FREQUENCY', '출현 빈도', '각 번호별 역대 당첨 횟수 기반 추천', 'frequency', 1),
('REAPPEAR_PROB', '재등장 확률', '이전 회차 번호가 다음 회차에 다시 나올 확률 기반', 'probability', 2),
('FIRST_POSITION', '첫번째 위치', '첫번째 번호(Num1)로 자주 나오는 번호 추천', 'position', 3),
('LAST_POSITION', '마지막 위치', '마지막 번호(Num6)로 자주 나오는 번호 추천', 'position', 4),
('PAIR_FREQUENCY', '동반 출현', '함께 자주 나오는 번호 쌍 기반 추천', 'pattern', 5),
('CONSECUTIVE', '연번 패턴', '연속 번호 패턴 기반 추천', 'pattern', 6),
('ODD_EVEN_RATIO', '홀짝 비율', '최적 홀짝 비율(3:3) 기반 추천', 'pattern', 7),
('HIGH_LOW_RATIO', '고저 비율', '최적 고저 비율(3:3) 기반 추천', 'pattern', 8),
('BAYESIAN', '베이지안 분석', '베이지안 사후확률 기반 추천', 'probability', 9),
('HOT_COLD', 'HOT/COLD', '최근 자주 나온(HOT)/안 나온(COLD) 번호 조합', 'frequency', 10)
ON CONFLICT (code) DO NOTHING;
