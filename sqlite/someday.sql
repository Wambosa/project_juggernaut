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
	LastUpdated		TIMESTAMP DEFAULT CURRENT_TIMESTAMP -- when this mind last checked in
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
INSERT INTO `ParcelType` VALUES
(1, 'Slack'),
(2, 'Lync'),
(3, 'SMTP'),
(4, 'SMS'),
(5, 'Voice'),
(6, 'FlowDock');

CREATE TABLE User ( -- a user can be used like a group as well
	UserId			INTEGER PRIMARY KEY AUTOINCREMENT,
	NickName		TEXT,
	FirstName		TEXT,
	LastName		TEXT,
	UserName		TEXT,		-- where we send the targeted message back	
	Area			TEXT,		-- where we send the targeted message back	
	ParcelTypeId	TEXT, 		-- where did we find this user?
	LastUpdated		TIMESTAMP	DEFAULT CURRENT_TIMESTAMP, -- ISO8601 YYYY-MM-DD HH:MM:SS.SSS

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
	CreatedOn		TIMESTAMP	DEFAULT CURRENT_TIMESTAMP, -- ISO8601 YYYY-MM-DD HH:MM:SS.SSS
	LastUpdated 	TIMESTAMP	DEFAULT CURRENT_TIMESTAMP,	-- ISO8601 YYYY-MM-DD HH:MM:SS.SSS

	FOREIGN KEY(JobStatusId) REFERENCES JobStatus(JobStatusId),
	FOREIGN KEY(ActionId) REFERENCES Action(ActionId),
	FOREIGN KEY(MindId) REFERENCES Mind(MindId),
	FOREIGN KEY(UserId) REFERENCES User(UserId)
);

CREATE TABLE JobData (
	JobDataId		INTEGER PRIMARY KEY AUTOINCREMENT,
	JobId			INTEGER,
	Key				TEXT,
	Value			TEXT,

	FOREIGN KEY(JobId) REFERENCES Job(JobId)
);

CREATE TABLE ParseStatus (
	ParseStatusId 	INTEGER PRIMARY KEY AUTOINCREMENT,
	ParseStatusName	TEXT
);
INSERT INTO `ParseStatus` VALUES
(1, 'NEW'),			-- just received from parcel pirate
(2, 'STARTED'),		-- might not use this unless parsing starts taking a long time
(3, 'COMPLETE'),	-- then there is a new job
(4, 'FAIL');		-- then there is just a response record

CREATE TABLE ReceivedMessage (
	ReceivedMessageId 	INTEGER PRIMARY KEY AUTOINCREMENT,
	ParcelTypeId		INT,		-- where did this message come from
	MessageText			TEXT,		-- the raw received message
	UserId				INTEGER,	-- who owns this message
	JobId				INTEGER,	-- if parsing was successful
	ParseStatusId		INTEGER 	DEFAULT 1, -- lexy progression of parsing
	CreatedOn			TIMESTAMP	DEFAULT CURRENT_TIMESTAMP, -- ISO8601 YYYY-MM-DD HH:MM:SS.SSS
	LastUpdated 		TIMESTAMP	DEFAULT CURRENT_TIMESTAMP, -- ISO8601 YYYY-MM-DD HH:MM:SS.SSS

	FOREIGN KEY(ParcelTypeId) REFERENCES ParcelType(ParcelTypeId),
	FOREIGN KEY(UserId) REFERENCES User(UserId),
	FOREIGN KEY(JobId) REFERENCES Job(JobId),
	FOREIGN KEY(ParseStatusId) REFERENCES ParseStatus(ParseStatusId)
);

CREATE TABLE ResponseMessage (
	ResponseMessageId 	INTEGER PRIMARY KEY AUTOINCREMENT,
	ParcelTypeId		INT,		-- how to send this message
	MessageText			TEXT,		-- what will be sent back to user
	JobId				INT,		-- the job that was created as a result of this message
	MindId				INTEGER,	-- who made this repsonse (0 = lexy)
	UserId				INTEGER,	-- who gets this message	
	CreatedOn			TIMESTAMP	DEFAULT CURRENT_TIMESTAMP, -- ISO8601 YYYY-MM-DD HH:MM:SS.SSS
	LastUpdated 		TIMESTAMP	DEFAULT CURRENT_TIMESTAMP,	-- ISO8601 YYYY-MM-DD HH:MM:SS.SSS
	IsDelivered			BOOLEAN,	-- returnosarus checks cares about this

	FOREIGN KEY(MindId) REFERENCES Mind(MindId),
	FOREIGN KEY(JobId) REFERENCES Job(JobId),
	FOREIGN KEY(UserId) REFERENCES User(UserId)
);

CREATE TABLE MindVocabulary (--TODO: potentially rename to TriggerPhrase
	MindVocabularyId	INTEGER PRIMARY KEY AUTOINCREMENT,
	SimpleWord			TEXT,	-- the word that the machine knows and can translate to a command
	Regex				TEXT,	-- the match that my machine must make
	Certainty			INTEGER	-- the amount of points that should be added towards a particular interpretation
);
INSERT INTO `MindVocabulary` (`MindVocabularyId`, `SimpleWord`, `Regex`, `Certainty`) VALUES
(1, 'define', 'define\b', 75),
(2, 'define', 'definition\b|defnition\b', 100),
(3, 'define', 'meaning\sof\b', 50),
(4, 'define', 'dictionary', 50),
(5, 'define', 'what\sis', 25),
(6, 'joke', 'joke\b', 100);

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
	MindVocabularyId		INTEGER NOT NULL,
	ParameterReasoningId	INTEGER NOT NULL,

	FOREIGN KEY(MindVocabularyId) REFERENCES MindVocabulary(MindVocabularyId),
	FOREIGN KEY(ParameterReasoningId) REFERENCES ParameterReasoning(ParameterReasoningId)
);
INSERT INTO `ParameterReasoningGrouping` (`ParameterName`, `MindVocabularyId`, `ParameterReasoningId`) VALUES
('wordToDefine', 1, 2),
('wordToDefine', 2, 1),
('wordToDefine', 3, 2),
('wordToDefine', 4, 1),
('wordToDefine', 5, 2);

--
--
-- test data only
INSERT INTO ReceivedMessage (ParcelTypeId, MessageText) VALUES
(1, 'define fool'),
(1, 'define fool for me'),
(1, 'define fool you fool'),
(1, 'define the word fool'),
(1, 'how does the dictionary describe fool'),
(1, 'what is a fool?'),
(1, 'what is the meaning of fool?'),
(1, 'meaning of fool?'),
(1, 'fool definition'),
(1, 'jew definition of fool'),
(1, 'i like blades');

COMMIT;