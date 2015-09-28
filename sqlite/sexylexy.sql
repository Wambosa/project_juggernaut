BEGIN TRANSACTION;

PRAGMA foreign_keys = 1;

CREATE TABLE TriggerPhrase (
	TriggerPhraseId		INTEGER PRIMARY KEY AUTOINCREMENT,
	ActionName			TEXT,	-- the word that the machine knows and can translate to a command
	Regex				TEXT,	-- the match that my machine must make
	Certainty			INTEGER	-- the amount of points that should be added towards a particular interpretation
);
INSERT INTO `TriggerPhrase` (`TriggerPhraseId`, `ActionName`, `Regex`, `Certainty`) VALUES
(1, 'Define', 'define\b', 75),
(2, 'Define', 'definition\b|defnition\b', 100),
(3, 'Define', 'meaning\sof\b', 50),
(4, 'Define', 'dictionary', 50),
(5, 'Define', 'what\sis', 25),
(6, 'Joke', 'chuck\b', 10),
(7, 'Joke', 'joke\b', 100),
(8, 'CannedResponse', 'i\sneed\shelp\b', 100), -- try last word as param logic
(9, 'CannedResponse', 'i\swant\shelp\b', 100), -- the word will determine the kind of response
(10, 'CannedResponse', 'i\swant\shelp\b', 100), -- helpdesk noob for the word help at the end
(11, 'CannedResponse', 'i\b.*help\b', 1), -- a attempted catch all
(12, 'CannedResponse', 'lol', 1),
(13, 'CannedResponse', 'fool\b', 1), -- canned responses should support a variant by default
(14, 'DirectExec', 'DirectExec:\b', 150), --asking user needs key AllowDirectExec val 1
;

CREATE TABLE ParameterReasoning (
	ParameterReasoningId	INTEGER PRIMARY KEY AUTOINCREMENT,
	Description				TEXT,	-- what was i thinking when i made this row
	IsIncludeOriginRegex	BOOLEAN,-- do we need to include the regex string in MindVocabulary
	Method					TEXT,	-- exec test match search replace split
	Regex					TEXT	-- the match that will give us the argument
);
INSERT INTO `ParameterReasoning` VALUES
(1, 'the very last word in the message becomes the argument data', 0, 'match', '\w{1,}(?:.(?!\w{1,}))+$'),
(2, 'the word immediately after vocabulary word becomes the argument', 1, 'match', '(?:{0}\s)\w{1,}');

CREATE TABLE ParameterReasoningGrouping (
	ParameterGroupingId		INTEGER PRIMARY KEY AUTOINCREMENT,
	ParameterName			TEXT, --becomes the key in key/value tables
	TriggerPhraseId			INTEGER NOT NULL,
	ParameterReasoningId	INTEGER NOT NULL,

	FOREIGN KEY(TriggerPhraseId) REFERENCES TriggerPhrase(TriggerPhraseId),
	FOREIGN KEY(ParameterReasoningId) REFERENCES ParameterReasoning(ParameterReasoningId)
);
INSERT INTO `ParameterReasoningGrouping` (`ParameterName`, `TriggerPhraseId`, `ParameterReasoningId`) VALUES
('WordToDefine', 1, 2),
('WordToDefine', 2, 1),
('WordToDefine', 3, 2),
('WordToDefine', 4, 1),
('WordToDefine', 5, 2);

COMMIT;