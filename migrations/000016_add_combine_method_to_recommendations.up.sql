-- lotto_recommendations 테이블에 조합 방법 컬럼 추가
ALTER TABLE lotto_recommendations ADD COLUMN IF NOT EXISTS combine_method VARCHAR(20) DEFAULT 'SIMPLE_AVG';
