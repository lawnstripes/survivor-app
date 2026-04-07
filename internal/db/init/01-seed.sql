-- Seed Users
INSERT OR IGNORE INTO users (name, slug, image_url, avatar_url, is_admin) VALUES
('John', 'john', '/static/users/john.png', '/static/users/avatars/john.png',0),
('Shawn', 'shawn', '/static/users/shawn.png', '/static/users/avatars/shawn.png',0),
('Matt', 'matt', '/static/users/matt.png', '/static/users/avatars/matt.png',0),
('Jen', 'jen', '/static/users/jen.png', '/static/users/avatars/jen.png',0),
('Alex', 'alex', '/static/users/alex.png', '/static/users/avatars/alex.png',1);

-- Seed Seasons
INSERT OR IGNORE INTO seasons (season, year, title, title_long, episodes, logo_url) VALUES
(50, 2026, 'Survivor 50', 'Survivor 50: In the Hands of the Fans', 13, '/static/seasons/survivor_50_logo.png'),
(49, 2025, 'Survivor 49', 'Survivor 49', 13, '/static/seasons/survivor_49_logo.png'),
(48, 2025, 'Survivor 48', 'Survivor 48', 12, '/static/seasons/survivor_48_logo.png'),
(47, 2024, 'Survivor 47', 'Survivor 47', 14, '/static/seasons/survivor_48_logo.png');


-- Seed AppData
INSERT OR IGNORE INTO appdata (id, draft_active, current_season)
SELECT 0, 0, s.id FROM seasons s WHERE s.season = 50;

-- Seed Contestants for Season 50
WITH c (name, known_by, image_url, elimination_episode, added_episode, is_winner) AS 
(
    VALUES
        ('Jenna Lewis-Dougherty', 'Jenna', '/static/contestants/s50_jenna.jpg', 1, 1, 0),
        ('Kyle Fraser', 'Kyle', '/static/contestants/s50_kyle.jpg', 1, 1, 0),
        ('Savannah Louie', 'Savannah', '/static/contestants/s50_savannah.jpg', 2, 1, 0),
        ('Q Burdette', 'Q', '/static/contestants/s50_q.jpg', 3, 1, 0),
        ('Mike White', 'Mike', '/static/contestants/s50_mike.jpg', 4, 1, 0),
        ('Charlie Davis', 'Charlie', '/static/contestants/s50_charlie.jpg', 5, 1, 0),
        ('Angelina Keeley', 'Angelina', '/static/contestants/s50_angelina.jpg', 5, 4, 0),
        ('Genevieve Mushaluk', 'Genevieve', '/static/contestants/s50_genevieve.jpg', 6, 1, 0),
        ('Colby Donaldson', 'Colby', '/static/contestants/s50_colby.jpg', 6, 2, 0),
        ('Kamilla Karthigesu', 'Kamilla', '/static/contestants/s50_kamilla.jpg', 6, 3, 0),
        ('Dee Valladares', 'Dee', '/static/contestants/s50_dee.jpg', 7, 1, 0),
        ('Stephenie LaGrossa', 'Steph', '/static/contestants/s50_stephenie.jpg', -1, 1, 0),
        ('Ozzy Lusth', 'Ozzy', '/static/contestants/s50_ozzy.jpg', -1, 1, 0),
        ('Benjamin Wade', 'Coach', '/static/contestants/s50_coach.jpg', 8, 1, 0),
        ('Aubry Bracco', 'Aubry', '/static/contestants/s50_aubry.jpg', -1, 1, 0),
        ('Chrissy Hoffbeck', 'Chrissy', '/static/contestants/s50_chrissy.jpg', 8, 1, 0),
        ('Christian Hubicki', 'Christian', '/static/contestants/s50_christian.jpg', -1, 1, 0),
        ('Jonathan Young', 'Jonathan', '/static/contestants/s50_jonathan.jpg', -1, 1, 0),
        ('Emily Flippen', 'Emily', '/static/contestants/s50_emily.jpg', -1, 1, 0),
        ('Tiffany Ervin', 'Tiffany', '/static/contestants/s50_tiffany.jpg', -1, 1, 0),
        ('Joe Hunter', 'Joe', '/static/contestants/s50_joe.jpg', -1, 1, 0),
        ('Cirie Fields', 'Cirie', '/static/contestants/s50_cirie.jpg', -1, 1, 0),
        ('Rick Devins', 'Devins', '/static/contestants/s50_rick.jpg', -1, 1, 0),
        ('Rizo Velovic', 'Rizo', '/static/contestants/s50_rizo.jpg', -1, 1, 0)
) 
INSERT OR IGNORE INTO contestants (season_id, name, known_by, image_url, elimination_episode, added_episode, is_winner)
SELECT
    s.id,
    c.name,
    c.known_by,
    c.image_url,
    c.elimination_episode,
    c.added_episode,
    c.is_winner
FROM c
JOIN seasons s on s.Season = 50;

-- Assign Teams for Season 50
UPDATE contestants SET owner_id = (SELECT id FROM users WHERE slug = 'john') WHERE season_id = (SELECT id FROM seasons WHERE season = 50) AND known_by IN ('Aubry', 'Cirie', 'Christian', 'Devins');
UPDATE contestants SET owner_id = (SELECT id FROM users WHERE slug = 'jen') WHERE season_id = (SELECT id FROM seasons WHERE season = 50) AND known_by IN ('Joe', 'Rizo', 'Genevieve', 'Tiffany');
UPDATE contestants SET owner_id = (SELECT id FROM users WHERE slug = 'alex') WHERE season_id = (SELECT id FROM seasons WHERE season = 50) AND known_by IN ('Kyle', 'Savannah', 'Charlie', 'Emily', 'Colby', 'Kamilla');
UPDATE contestants SET owner_id = (SELECT id FROM users WHERE slug = 'shawn') WHERE season_id = (SELECT id FROM seasons WHERE season = 50) AND known_by IN ('Dee', 'Jonathan', 'Coach', 'Steph');
UPDATE contestants SET owner_id = (SELECT id FROM users WHERE slug = 'matt') WHERE season_id = (SELECT id FROM seasons WHERE season = 50) AND known_by IN ('Chrissy', 'Mike', 'Ozzy', 'Q', 'Angelina');

WITH c(name, known_by, image_url, elimination_episode, added_episode, is_winner) AS
(
        VALUES 		
        ('Alex More', 'Alex', '/static/contestants/s49_alex.jpg', 9, 1, 0),
		('Kimberly Davis', 'Annie', '/static/contestants/s49_annie.jpg', 2, 1, 0),
		('Jake Latimer', 'Jake', '/static/contestants/s49_jake.jpg', 3, 1, 0),
		('Jason Treul', 'Jason', '/static/contestants/s49_jason.jpg', 5, 1, 0),
		('Jawan Pitts', 'Jawan', '/static/contestants/s49_jawan.jpg', 10,1, 0),
		('Jeremiah Ing', 'Jeremiah', '/static/contestants/s49_jeremiah.jpg', 3,1, 0),
		('Kristina Mills', 'Kristina', '/static/contestants/s49_kristina.jpg', 13, 4, 0),
		('Matt Williams', 'Matt', '/static/contestants/s49_matt.jpg', 4,1, 0),
		('Michelle Chukwujekwu', 'MC', '/static/contestants/s49_mc.jpg', 8, 2, 0),
		('Nate Moore', 'Nate', '/static/contestants/s49_nate.jpg', 7,  1, 0),
		('Nicole Mazullo', 'Nicole', '/static/contestants/s49_nicole.jpg', 1, 1, 0),
		('Rizo Velovic', 'Rizo', '/static/contestants/s49_rizo.jpg', 13, 1, 0),
		('Sage Ahrens-Nicols', 'Sage', '/static/contestants/s49_sage.jpg',-1, 1, 0),
		('Savannah Louie', 'Savannah', '/static/contestants/s49_savannah.jpg',-1, 1, 1),
		('Shannon Fairweather', 'Shannon', '/static/contestants/s49_shannon.jpg', 6,1, 0),
		('Sophi Balerdi', 'Sophi', '/static/contestants/s49_sophi.jpg', 11,1, 0),
		('Sophie Segreti', 'Sophie', '/static/contestants/s49_sophie.jpg',-1,1, 0),
		('Steven Ramm', 'Steven', '/static/contestants/s49_steven.jpg', 12,1,0)
) 
INSERT OR IGNORE INTO contestants (season_id, name, known_by, image_url, elimination_episode, added_episode, is_winner)
SELECT
    s.id,
    c.name,
    c.known_by,
    c.image_url,
    c.elimination_episode,
    c.added_episode,
    c.is_winner
FROM c
JOIN seasons s on s.Season = 49;

UPDATE contestants SET owner_id = (SELECT id FROM users WHERE slug = 'john') WHERE season_id = (SELECT id FROM seasons WHERE season = 49) AND known_by IN ('Savannah', 'Rizo', 'Sophie');
UPDATE contestants SET owner_id = (SELECT id FROM users WHERE slug = 'jen') WHERE season_id = (SELECT id FROM seasons WHERE season = 49) AND known_by IN ('Sage', 'Jason', 'Matt');
UPDATE contestants SET owner_id = (SELECT id FROM users WHERE slug = 'alex') WHERE season_id = (SELECT id FROM seasons WHERE season = 49) AND known_by IN ('Nate', 'Jeremiah', 'Nicole', 'MC');
UPDATE contestants SET owner_id = (SELECT id FROM users WHERE slug = 'shawn') WHERE season_id = (SELECT id FROM seasons WHERE season = 49) AND known_by IN ('Jake', 'Shannon', 'Sophi', 'Kristina');
UPDATE contestants SET owner_id = (SELECT id FROM users WHERE slug = 'matt') WHERE season_id = (SELECT id FROM seasons WHERE season = 49) AND known_by IN ('Steven', 'Alex', 'Jawan');

WITH c (name, known_by, image_url, elimination_episode, added_episode, is_winner) AS
(
        VALUES 	
   	    ('Bianca Roses', 'Bianca', '/static/contestants/s48_bianca.jpg', 5, 1, 0),
		('Cedrek McFadden', 'Cedrek', '/static/contestants/s48_cedrek.jpg', 7, 1, 0),
		('Charity  Nelms', 'Charity', '/static/contestants/s48_charity.jpg', 6, 1, 0),
		('Chrissy Sarnowsky', 'Chrissy', '/static/contestants/s48_chrissy.jpg', 8, 1, 0),
		('David Kinne', 'David', '/static/contestants/s48_david.jpg', 9, 1, 0),
		('Eva Erickson', 'Eva', '/static/contestants/s48_eva.jpg', -1, 1, 0 ),
		('Joe Hunter', 'Joe', '/static/contestants/s48_joe.jpg', -1 , 1, 0),
		('Justin Pioppi', 'Justin', '/static/contestants/s48_justin.jpg', 3, 1, 0),
		('Kamilla Karthigesu', 'Kamilla', '/static/contestants/s48_kamilla.jpg', 13, 3, 0),
		('Kevin Leung', 'Kevin', '/static/contestants/s48_kevin.jpg', 2, 1, 0),
	   	('Kyle Fraser', 'Kyle', '/static/contestants/s48_kyle.jpg', -1, 1, 1),
		('Mary Zheng', 'Mary', '/static/contestants/s48_mary.jpg', 11, 1, 0),
		('Mitch Guerra', 'Mitch', '/static/contestants/s48_mitch.jpg', 13, 1, 0),
		('Shauhin Davari', 'Shauhin', '/static/contestants/s48_shauhin.jpg', 12, 1, 0),
		('Star Toomey', 'Star', '/static/contestants/s48_star.jpg', 10, 4, 0),
		('Stephanie Berger', 'Stephanie', '/static/contestants/s48_stephanie.jpg', 1, 1, 0),
		('Thomas Krottinger', 'Thomas', '/static/contestants/s48_thomas.jpg', 4, 1, 0),
		('Saiounia Hughley', 'Sai', '/static/contestants/s48_sai.jpg', 7, 1, 0)
) 
INSERT OR IGNORE INTO contestants (season_id, name, known_by, image_url, elimination_episode, added_episode, is_winner)
SELECT
    s.id,
    c.name,
    c.known_by,
    c.image_url,
    c.elimination_episode,
    c.added_episode,
    c.is_winner
FROM c
JOIN seasons s on s.Season = 48;
	
UPDATE contestants SET owner_id = (SELECT id FROM users WHERE slug = 'john') WHERE season_id = (SELECT id FROM seasons WHERE season = 48) AND known_by IN ('Shauhin', 'Joe', 'Bianca');
UPDATE contestants SET owner_id = (SELECT id FROM users WHERE slug = 'jen') WHERE season_id = (SELECT id FROM seasons WHERE season = 48) AND known_by IN ('Charity', 'David', 'Cedrek');
UPDATE contestants SET owner_id = (SELECT id FROM users WHERE slug = 'alex') WHERE season_id = (SELECT id FROM seasons WHERE season = 48) AND known_by IN ('Eva', 'Mitch', 'Thomas');
UPDATE contestants SET owner_id = (SELECT id FROM users WHERE slug = 'shawn') WHERE season_id = (SELECT id FROM seasons WHERE season = 48) AND known_by IN ('Kyle', 'Kevin', 'Sai', 'Kamilla');
UPDATE contestants SET owner_id = (SELECT id FROM users WHERE slug = 'matt') WHERE season_id = (SELECT id FROM seasons WHERE season = 48) AND known_by IN ('Justin', 'Mary', 'Chrissy', 'Star');

WITH c (name, known_by, image_url, elimination_episode, added_episode, is_winner) AS
(
VALUES 
('Anika Dhar','Anika','/static/contestants/s47_anika.jpg',5,0,0),
('Andy Rueda','Andy','/static/contestants/s47_andy.jpg',13,0,0),
('Aysha Welsh','Aysha','/static/contestants/s47_aysha.jpg',3,0,0),
('Caroline Vidmar','Caroline','/static/contestants/s47_caroline.jpg',12,1,0),
('Gabe Ortis','Gabe','/static/contestants/s47_gabe.jpg',10,1,0),
('Genevieve Mushaluk','Genevieve','/static/contestants/s47_genevieve.jpg',13,1,0),
('Jon Lovett','Jon','/static/contestants/s47_jon.jpg',1,1,0),
('Kishan Patel','Kishan','/static/contestants/s47_kishan.jpg',4,1,0),
('Kyle Ostwald','Kyle','/static/contestants/s47_kyle.jpg',11,1,0),
('Rachel Lamont','Rachel','/static/contestants/s47_rachel.jpg',0,3,1),
('Rome Cooney','Rome','/static/contestants/s47_rome.jpg',6,1,0),
('Sam Phalen','Sam','/static/contestants/s47_sam.jpg',0,1,0),
('Sierra Wright','Sierra','/static/contestants/s47_sierra.jpg',8,1,0),
('Solomon Yi','Sol','/static/contestants/s47_sol.jpg',9,1,0),
('Sue Smey','Sue','/static/contestants/s47_sue.jpg',0,1,0),
('Teeny Chirichillo','Teeny','/static/contestants/s47_teeny.jpg',14,2,0),
('Terran Foster','TK','/static/contestants/s47_tk.jpg',2,1,0),
('Tiyana Hallums','Tiyana','/static/contestants/s47_tiyana.jpg',7,1,0)
) 
INSERT OR IGNORE INTO contestants (season_id, name, known_by, image_url, elimination_episode, added_episode, is_winner)
SELECT
    s.id,
    c.name,
    c.known_by,
    c.image_url,
    c.elimination_episode,
    c.added_episode,
    c.is_winner
FROM c
JOIN seasons s on s.Season = 47;

UPDATE contestants SET owner_id = (SELECT id FROM users WHERE slug = 'john') WHERE season_id = (SELECT id FROM seasons WHERE season = 47) AND known_by IN ('Andy','Jon','Teeny','Tiyana');
UPDATE contestants SET owner_id = (SELECT id FROM users WHERE slug = 'jen') WHERE season_id = (SELECT id FROM seasons WHERE season = 47) AND known_by IN ('Kyle','Rome','Sue');
UPDATE contestants SET owner_id = (SELECT id FROM users WHERE slug = 'alex') WHERE season_id = (SELECT id FROM seasons WHERE season = 47) AND known_by IN ('Caroline','Gabe','Kishan');
UPDATE contestants SET owner_id = (SELECT id FROM users WHERE slug = 'shawn') WHERE season_id = (SELECT id FROM seasons WHERE season = 47) AND known_by IN ('Rachel','Sierra','Sol','TK');
UPDATE contestants SET owner_id = (SELECT id FROM users WHERE slug = 'matt') WHERE season_id = (SELECT id FROM seasons WHERE season = 47) AND known_by IN ('Anika','Genevieve','Sam');