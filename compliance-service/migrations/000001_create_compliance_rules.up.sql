-- =============================================================================
-- Migration: Create compliance_rules table
-- =============================================================================

-- Create enum types
CREATE TYPE rule_severity AS ENUM ('INFO', 'WARNING', 'ERROR');
CREATE TYPE rule_category AS ENUM (
    'load_bearing',
    'wet_zones', 
    'fire_safety',
    'ventilation',
    'min_area',
    'daylight',
    'accessibility',
    'general'
);
CREATE TYPE approval_type AS ENUM (
    'NONE',
    'NOTIFICATION',
    'PROJECT',
    'EXPERTISE',
    'PROHIBITED'
);

-- Create compliance_rules table
CREATE TABLE IF NOT EXISTS compliance_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(100) NOT NULL UNIQUE,
    category rule_category NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    severity rule_severity NOT NULL DEFAULT 'WARNING',
    active BOOLEAN NOT NULL DEFAULT true,
    applies_to JSONB NOT NULL DEFAULT '[]',
    applies_to_operations JSONB NOT NULL DEFAULT '[]',
    approval_required approval_type NOT NULL DEFAULT 'NONE',
    parameters JSONB NOT NULL DEFAULT '{}',
    "references" JSONB NOT NULL DEFAULT '[]',
    version VARCHAR(20) NOT NULL DEFAULT '1.0.0',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create indexes
CREATE INDEX idx_compliance_rules_category ON compliance_rules(category);
CREATE INDEX idx_compliance_rules_severity ON compliance_rules(severity);
CREATE INDEX idx_compliance_rules_active ON compliance_rules(active);
CREATE INDEX idx_compliance_rules_applies_to ON compliance_rules USING GIN (applies_to);
CREATE INDEX idx_compliance_rules_applies_to_operations ON compliance_rules USING GIN (applies_to_operations);

-- Create updated_at trigger
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_compliance_rules_updated_at
    BEFORE UPDATE ON compliance_rules
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- =============================================================================
-- Seed initial compliance rules (СНиП и ЖК РФ)
-- =============================================================================

-- 1. Несущие конструкции
INSERT INTO compliance_rules (code, category, name, description, severity, applies_to, applies_to_operations, approval_required, parameters, "references") VALUES
(
    'SNIP-31-01-2003-9.22',
    'load_bearing',
    'Запрет сноса несущих стен',
    'Снос несущих стен и конструкций запрещён без проекта и экспертизы. Несущие стены обеспечивают устойчивость здания.',
    'ERROR',
    '["LOAD_BEARING_WALL"]',
    '["DEMOLISH_WALL"]',
    'EXPERTISE',
    '{}',
    '[{"code": "СНиП 31-01-2003", "title": "Здания жилые многоквартирные", "section": "п.9.22"}]'
),
(
    'SNIP-31-01-2003-9.23',
    'load_bearing',
    'Устройство проёмов в несущих стенах',
    'Устройство проёмов в несущих стенах допускается только по проекту с усилением конструкций.',
    'ERROR',
    '["LOAD_BEARING_WALL"]',
    '["ADD_OPENING"]',
    'EXPERTISE',
    '{"max_opening_width": 0.9}',
    '[{"code": "СНиП 31-01-2003", "title": "Здания жилые многоквартирные", "section": "п.9.23"}]'
);

-- 2. Мокрые зоны
INSERT INTO compliance_rules (code, category, name, description, severity, applies_to, applies_to_operations, approval_required, parameters, "references") VALUES
(
    'ZHK-RF-25',
    'wet_zones',
    'Запрет размещения санузлов над жилыми комнатами',
    'Санузлы и ванные не могут располагаться над жилыми комнатами и кухнями нижерасположенных квартир.',
    'ERROR',
    '["BATHROOM", "TOILET", "WET_ZONE"]',
    '["MOVE_WET_ZONE", "EXPAND_WET_ZONE"]',
    'PROHIBITED',
    '{}',
    '[{"code": "ЖК РФ", "title": "Жилищный кодекс РФ", "section": "ст.25"}]'
),
(
    'SP-54.13330-5.8',
    'wet_zones',
    'Гидроизоляция мокрых зон',
    'При расширении мокрых зон требуется устройство гидроизоляции пола и стен.',
    'WARNING',
    '["BATHROOM", "TOILET", "WET_ZONE"]',
    '["EXPAND_WET_ZONE"]',
    'PROJECT',
    '{}',
    '[{"code": "СП 54.13330.2016", "title": "Здания жилые многоквартирные", "section": "п.5.8"}]'
),
(
    'SNIP-KITCHEN-VENT',
    'wet_zones',
    'Кухня должна иметь вытяжную вентиляцию',
    'Кухня должна быть оборудована вытяжной вентиляцией с естественным или механическим побуждением.',
    'ERROR',
    '["KITCHEN"]',
    '["CHANGE_ROOM_TYPE"]',
    'PROJECT',
    '{}',
    '[{"code": "СП 54.13330.2016", "title": "Здания жилые многоквартирные", "section": "п.9.7"}]'
);

-- 3. Минимальные площади
INSERT INTO compliance_rules (code, category, name, description, severity, applies_to, applies_to_operations, approval_required, parameters, "references") VALUES
(
    'SP-54-MIN-LIVING',
    'min_area',
    'Минимальная площадь жилой комнаты',
    'Площадь жилой комнаты должна быть не менее 8 кв.м для однокомнатных квартир и 10 кв.м для многокомнатных.',
    'ERROR',
    '["ROOM"]',
    '["SPLIT_ROOM", "MERGE_ROOMS"]',
    'NONE',
    '{"min_area_single": 8.0, "min_area_multi": 10.0}',
    '[{"code": "СП 54.13330.2016", "title": "Здания жилые многоквартирные", "section": "п.5.7"}]'
),
(
    'SP-54-MIN-KITCHEN',
    'min_area',
    'Минимальная площадь кухни',
    'Площадь кухни должна быть не менее 5 кв.м.',
    'ERROR',
    '["KITCHEN"]',
    '["SPLIT_ROOM"]',
    'NONE',
    '{"min_area": 5.0}',
    '[{"code": "СП 54.13330.2016", "title": "Здания жилые многоквартирные", "section": "п.5.7"}]'
),
(
    'SP-54-MIN-BATHROOM',
    'min_area',
    'Минимальная площадь санузла',
    'Площадь совмещённого санузла должна быть не менее 3.8 кв.м, раздельного туалета - 1.2 кв.м.',
    'WARNING',
    '["BATHROOM", "TOILET"]',
    '["SPLIT_ROOM"]',
    'NONE',
    '{"min_area_combined": 3.8, "min_area_toilet": 1.2, "min_area_bathroom": 2.5}',
    '[{"code": "СП 54.13330.2016", "title": "Здания жилые многоквартирные", "section": "п.5.7"}]'
);

-- 4. Вентиляция
INSERT INTO compliance_rules (code, category, name, description, severity, applies_to, applies_to_operations, approval_required, parameters, "references") VALUES
(
    'SP-54-VENT-CHANNELS',
    'ventilation',
    'Запрет переноса вентиляционных каналов',
    'Перенос, уменьшение сечения или демонтаж вентиляционных каналов запрещён.',
    'ERROR',
    '["VENTILATION"]',
    '["MOVE_VENTILATION"]',
    'PROHIBITED',
    '{}',
    '[{"code": "СП 54.13330.2016", "title": "Здания жилые многоквартирные", "section": "п.9.6"}]'
),
(
    'SP-54-VENT-ACCESS',
    'ventilation',
    'Доступ к вентиляционным каналам',
    'Должен быть обеспечен доступ к вентиляционным каналам для обслуживания.',
    'WARNING',
    '["VENTILATION"]',
    '["ADD_WALL"]',
    'NONE',
    '{"min_access_distance": 0.3}',
    '[{"code": "СП 54.13330.2016", "title": "Здания жилые многоквартирные", "section": "п.9.6"}]'
);

-- 5. Пожарная безопасность
INSERT INTO compliance_rules (code, category, name, description, severity, applies_to, applies_to_operations, approval_required, parameters, "references") VALUES
(
    'SP-1-MIN-DOOR-WIDTH',
    'fire_safety',
    'Минимальная ширина дверных проёмов',
    'Ширина дверных проёмов эвакуационных выходов должна быть не менее 0.8 м.',
    'ERROR',
    '["DOOR"]',
    '["ADD_OPENING", "CLOSE_OPENING"]',
    'NONE',
    '{"min_width": 0.8}',
    '[{"code": "СП 1.13130.2009", "title": "Системы противопожарной защиты", "section": "п.4.2.5"}]'
),
(
    'SP-1-SECOND-EXIT',
    'fire_safety',
    'Второй эвакуационный выход',
    'Квартиры с площадью более 100 кв.м должны иметь не менее двух эвакуационных выходов.',
    'WARNING',
    '["ROOM"]',
    '["CLOSE_OPENING"]',
    'PROJECT',
    '{"max_area_single_exit": 100.0}',
    '[{"code": "СП 1.13130.2009", "title": "Системы противопожарной защиты", "section": "п.5.2.9"}]'
);

-- 6. Естественное освещение
INSERT INTO compliance_rules (code, category, name, description, severity, applies_to, applies_to_operations, approval_required, parameters, "references") VALUES
(
    'SP-54-DAYLIGHT',
    'daylight',
    'Естественное освещение жилых комнат',
    'Жилые комнаты должны иметь естественное освещение (окна).',
    'ERROR',
    '["ROOM"]',
    '["CHANGE_ROOM_TYPE", "CLOSE_OPENING"]',
    'NONE',
    '{"window_area_ratio": 0.125}',
    '[{"code": "СП 54.13330.2016", "title": "Здания жилые многоквартирные", "section": "п.9.12"}]'
),
(
    'SP-54-KITCHEN-LIGHT',
    'daylight',
    'Освещение кухни',
    'Кухня должна иметь естественное освещение или быть кухней-нишей площадью не более 5 кв.м.',
    'WARNING',
    '["KITCHEN"]',
    '["CHANGE_ROOM_TYPE", "CLOSE_OPENING"]',
    'NONE',
    '{"max_kitchen_niche_area": 5.0}',
    '[{"code": "СП 54.13330.2016", "title": "Здания жилые многоквартирные", "section": "п.5.7"}]'
);

-- 7. Общие правила
INSERT INTO compliance_rules (code, category, name, description, severity, applies_to, applies_to_operations, approval_required, parameters, "references") VALUES
(
    'ZHK-RF-26',
    'general',
    'Уведомление о перепланировке',
    'Перепланировка, не затрагивающая несущие конструкции, требует уведомления органов местного самоуправления.',
    'INFO',
    '["WALL", "ROOM"]',
    '["DEMOLISH_WALL", "ADD_WALL", "MERGE_ROOMS", "SPLIT_ROOM"]',
    'NOTIFICATION',
    '{}',
    '[{"code": "ЖК РФ", "title": "Жилищный кодекс РФ", "section": "ст.26"}]'
),
(
    'SP-54-CEILING-HEIGHT',
    'general',
    'Минимальная высота потолков',
    'Высота потолков в жилых комнатах и кухнях должна быть не менее 2.5 м (для мансард - 2.3 м).',
    'ERROR',
    '["ROOM", "KITCHEN"]',
    '[]',
    'NONE',
    '{"min_height": 2.5, "min_height_mansard": 2.3}',
    '[{"code": "СП 54.13330.2016", "title": "Здания жилые многоквартирные", "section": "п.5.8"}]'
);

