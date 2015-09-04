BEGIN TRANSACTION;

PRAGMA foreign_keys = 1;

CREATE TABLE Action (
	ActionId		INTEGER PRIMARY KEY AUTOINCREMENT,
	ActionName		TEXT	-- easy name (possible to have duplicates due to potentially different methods)
);
INSERT INTO `Action` VALUES
(1, 'define'),
(2, 'setBuildId');

CREATE TABLE ChainCommand (
	ChainCommandId			INTEGER PRIMARY KEY AUTOINCREMENT,
	ActionId				INTEGER,	-- the overall action desired
	Step					INTEGER,	-- used for ordering
	ExecText				TEXT,		-- what is executed on the command line. think string.Format
	FOREIGN KEY(ActionId) REFERENCES Action(ActionId)
);
INSERT INTO `ChainCommand` (`ActionId`, `Step`, `ExecText`) VALUES
(1, 1, 'echo someResultSetOrAnswerString{0}');

CREATE TABLE Mind (
	MindId			INTEGER PRIMARY KEY AUTOINCREMENT,
	MindName		TEXT,	-- friendly name that this mind responds to
	Nosiness		INTEGER,-- 1 - 100% chance of butting in conversation
	Sassyness		INTEGER,-- 1 - 100% chance of insulting user
	Speed			INTEGER,-- 5 - 600 seconds polling interval 
	UniqueAddress	TEXT,	-- MAC + first init epoch timestamp
	LastUpdated		TEXT	-- when this mind last checked in
);

CREATE TABLE MindCapability (
	MindCapabilityId 	INTEGER PRIMARY KEY AUTOINCREMENT,
	MindId				INTEGER,
	ActionId			INTEGER,

	FOREIGN KEY(MindId) REFERENCES Mind(MindId),
	FOREIGN KEY(ActionId) REFERENCES Action(ActionId)
);

CREATE TABLE ParcelType (
	ParcelTypeId 	INTEGER PRIMARY KEY AUTOINCREMENT,
	ParcelTypeName	TEXT -- what do we know this as
);
INSERT INTO `ParcelType` VALUES (1, 'Slack Api');
INSERT INTO `ParcelType` VALUES (2, 'Lync Client');
INSERT INTO `ParcelType` VALUES (3, 'SMTP Client');
INSERT INTO `ParcelType` VALUES (4, 'Google SMS');
INSERT INTO `ParcelType` VALUES (5, 'Google Voice');

CREATE TABLE User ( -- a user can be used like a group as well
	UserId			INTEGER PRIMARY KEY AUTOINCREMENT,
	NickName		TEXT,
	FirstName		TEXT,
	LastName		TEXT,
	UserName		TEXT, -- where we send the targeted message back	
	Area			TEXT, -- where we send the targeted message back	
	ParcelTypeId	TEXT, -- where did we find this user?
	LastUpdated		TEXT, -- ISO8601 YYYY-MM-DD HH:MM:SS.SSS

	FOREIGN KEY(ParcelTypeId) REFERENCES ParcelType(ParcelTypeId)
);

CREATE TABLE UserPreference (
	UserPreferenceId	INTEGER PRIMARY KEY AUTOINCREMENT,
	UserId				INTEGER,
	Key					TEXT,
	Value				TEXT,

	FOREIGN KEY(UserId) REFERENCES User(UserId)	
);
-- INSERT INTO `UserPreference` VALUES (1, 1, 'nosiness', '90');
-- this means that mind(s) will but into this users convos more often

CREATE TABLE JobStatus (
	JobStatusId 	INTEGER PRIMARY KEY AUTOINCREMENT,
	JobStatusName	TEXT
);
INSERT INTO `JobStatus` VALUES 
(1, 'NEW'),
(2, 'STARTED'),
(3, 'COMPLETE'),
(4, 'NEEDARGS'),		--ask the user for things
(5, 'BADARGS'),			--ask the user for things
(6, 'MISSINGMODULE'),	--this flags another mind to take over
(7, 'BROKENMODULE');	--this flags another mind to take over

CREATE TABLE Job (
	JobId 			INTEGER PRIMARY KEY AUTOINCREMENT,
	JobStatusId 	INTEGER,	-- bot progression of job
	ActionId		INTEGER,	-- action that should be taken
	Certainty		INTEGER,	-- the higher the number, the more certain i am that it is the right action
	MindId			INTEGER,	-- mind that has begun working on this
	UserId			INTEGER,	-- user that triggered this job
	CreatedOn		TEXT,		-- ISO8601 YYYY-MM-DD HH:MM:SS.SSS
	LastUpdated 	TEXT,		-- ISO8601 YYYY-MM-DD HH:MM:SS.SSS

	FOREIGN KEY(JobStatusId) REFERENCES JobStatus(JobStatusId),
	FOREIGN KEY(ActionId) REFERENCES Action(ActionId),
	FOREIGN KEY(MindId) REFERENCES Mind(MindId),
	FOREIGN KEY(UserId) REFERENCES User(UserId)
);

CREATE TABLE ParseStatus (
	ParseStatusId 	INTEGER PRIMARY KEY AUTOINCREMENT,
	ParseStatusName	TEXT
);
INSERT INTO `ParseStatus` VALUES (1, 'NEW');
INSERT INTO `ParseStatus` VALUES (2, 'STARTED');
INSERT INTO `ParseStatus` VALUES (3, 'COMPLETE'); -- then there is a new job
INSERT INTO `ParseStatus` VALUES (4, 'FAIL'); -- then there is just a response record

CREATE TABLE ReceivedMessage (--default val for parsestatus !! todo:
	ReceivedMessageId 	INTEGER PRIMARY KEY AUTOINCREMENT,
	ParcelTypeId		INT,		-- where did this message come from
	MessageText			TEXT,		-- the raw received message
	UserId				INTEGER,	-- who owns this message
	JobId				INTEGER,	-- if parsing was successful
	ParseStatusId		INTEGER,	-- lexy progression of parsing
	CreatedOn			TEXT,		-- ISO8601 YYYY-MM-DD HH:MM:SS.SSS
	LastUpdated 		TEXT,		-- ISO8601 YYYY-MM-DD HH:MM:SS.SSS

	FOREIGN KEY(ParcelTypeId) REFERENCES ParcelType(ParcelTypeId),
	FOREIGN KEY(UserId) REFERENCES User(UserId),
	FOREIGN KEY(JobId) REFERENCES Job(JobId),
	FOREIGN KEY(ParseStatusId) REFERENCES ParseStatus(ParseStatusId)
);

CREATE TABLE ResponseMessage (
	ResponseMessageId 	INTEGER PRIMARY KEY AUTOINCREMENT,
	ParcelTypeId		INT,		-- how to send this message
	MessageText			TEXT,		-- what will be sent back to user
	JobId				INT,		-- 
	MindId				INTEGER,	-- who made this repsonse (0 = lexy)
	UserId				INTEGER,	-- who gets this message	
	CreatedOn			TEXT,		-- ISO8601 YYYY-MM-DD HH:MM:SS.SSS
	LastUpdated 		TEXT,		-- ISO8601 YYYY-MM-DD HH:MM:SS.SSS
	IsDelivered			INTEGER,	-- bool

	FOREIGN KEY(MindId) REFERENCES Mind(MindId),
	FOREIGN KEY(JobId) REFERENCES Job(JobId),
	FOREIGN KEY(UserId) REFERENCES User(UserId)
);

CREATE TABLE MindVocabulary (
	MindVocabularyId	INTEGER PRIMARY KEY AUTOINCREMENT,
	SimpleWord			TEXT,	-- the word that the machine knows and can translate to a command
	Regex				TEXT,	-- the match that my machine must make
	Certainty			INTEGER	-- the amount of points that should be added towards a particular interpretation
);

--will iteratte over all possible regex and build up points for all unique simpleWords found. simple plus plus scoring system
INSERT INTO `MindVocabulary` (`SimpleWord`, `Regex`, `Certainty`) VALUES
('define', 'define\b', 75),
('define', 'definition\b', 100),
('define', 'meaning\sof\b', 50), --some regex with the same meaning or alternate mispellings can be joined to a single row with the regex OR clause. actually, only do this for mispellings
('define', 'dictionary', 50),
('define', 'what\sis', 25);


--test data
INSERT INTO ReceivedMessage
(ParcelTypeId, MessageText, ParseStatusId)
VALUES
(1, 'define fool', 1),
(1, 'define fool for me', 1),
(1, 'define fool you fool', 1),
(1, 'define the word fool', 1),
(1, 'how does the dictionary describe fool', 1),
(1, 'what is a fool?', 1),
(1, 'what is the meaning of fool?', 1),
(1, 'meaning of fool?', 1),
(1, 'fool definition', 1),
(1, 'jew definition of fool', 1),
(1, 'i like blades',1 );

COMMIT;