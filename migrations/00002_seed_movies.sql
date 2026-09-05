-- +goose Up
INSERT INTO movies (
    id,
    title,
    year,
    director
) VALUES
    (1, 'Побег из Шоушенка', 1994, 'Frank Darabont'),
    (2, 'Крестный отец', 1972, 'Francis Ford Coppola'),
    (3, 'Темный рыцарь', 2008, 'Christopher Nolan'),
    (4, 'Криминальное чтиво', 1994, 'Quentin Tarantino'),
    (5, 'Форрест Гамп', 1994, 'Robert Zemeckis'),
    (6, 'Начало', 2010, 'Christopher Nolan'),
    (7, 'Матрица', 1999, 'Lana Wachowski, Lilly Wachowski'),
    (8, 'Интерстеллар', 2014, 'Christopher Nolan'),
    (9, 'Унесённые призраками', 2001, 'Hayao Miyazaki'),
    (10, 'Паразиты', 2019, 'Bong Joon Ho'),
    (11, 'Гладиатор', 2000, 'Ridley Scott'),
    (12, 'Одержимость', 2014, 'Damien Chazelle'),
    (13, 'Назад в будущее', 1985, 'Robert Zemeckis'),
    (14, 'Чужой', 1979, 'Ridley Scott'),
    (15, 'Прибытие', 2016, 'Denis Villeneuve')
ON CONFLICT (id) DO NOTHING;

-- +goose Down
DELETE FROM movies WHERE id BETWEEN 1 AND 15;
