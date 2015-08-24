BEGIN TRANSACTION;

PRAGMA foreign_keys = 1;

CREATE TABLE Command (
	CommandId		INTEGER PRIMARY KEY AUTOINCREMENT,
	CommandName		TEXT,	-- easy name
	ExecText		TEXT	-- what is executed on the command line. think string.Format
);
INSERT INTO Command VALUES ('setBuildID', 'turtlebot.exe -getrevision "US00000" | turtlebot.exe -getlatestbuild "{0}" | rallybot.exe -buildId {0} {1}'


CREATE TABLE ChainCommand (
	ChainCommandId			INTEGER PRIMARY KEY AUTOINCREMENT,
	ChainCommandName		TEXT, 		-- name of joint command group
	Step					INTEGER,	--used for ordering
	CommandId				INTEGER		-- the command in this sequence

	FOREIGN KEY(CommandId) REFERENCES Command(CommandId)
);

INSERT INTO JointCommand VALUES (1, 'setBuildID', 1, 35); --turtlebot args
INSERT INTO JointCommand VALUES (2, 'setBuildID', 2, 35); --turtlebot args
INSERT INTO JointCommand VALUES (3, 'setBuildID', 3, 12); --rallybot


INSERT INTO JointCommand VALUES (4, 'setBuildID', 1, 35); --turtlebot args
INSERT INTO JointCommand VALUES (5, 'setBuildID', 2, 12); --rallybot


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
	CommandId			INTEGER,

	FOREIGN KEY(MindId) REFERENCES Mind(MindId),
	FOREIGN KEY(CommandId) REFERENCES Command(CommandId)
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
	JobStatudId 	INTEGER PRIMARY KEY AUTOINCREMENT,
	JobStatusName	TEXT
);
INSERT INTO `JobStatus` VALUES (1, 'NEW');
INSERT INTO `JobStatus` VALUES (2, 'STARTED');
INSERT INTO `JobStatus` VALUES (3, 'COMPLETE');
INSERT INTO `JobStatus` VALUES (4, 'NEEDARGS');		--ask the user for things
INSERT INTO `JobStatus` VALUES (5, 'BADARGS');		--ask the user for things
INSERT INTO `JobStatus` VALUES (6, 'MISSINGMODULE');--this flags another mind to take over
INSERT INTO `JobStatus` VALUES (7, 'BROKENMODULE');	--this flags another mind to take over

CREATE TABLE Job (
	JobId 		INTEGER PRIMARY KEY AUTOINCREMENT,
	JobStatusId INTEGER,	-- bot progression of job
	CommandId	INTEGER,	-- command that should be executed
	MindId		INTEGER,	-- mind that has begun working on this
	UserId		INTEGER,	-- user that triggered this job
	CreatedOn	TEXT,		-- ISO8601 YYYY-MM-DD HH:MM:SS.SSS
	LastUpdated TEXT,		-- ISO8601 YYYY-MM-DD HH:MM:SS.SSS

	FOREIGN KEY(JobStatudId) REFERENCES JobStatus(JobStatudId),
	FOREIGN KEY(CommandId) REFERENCES Command(CommandId),
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

CREATE TABLE RecievedMessage (
	RecievedMessageId 	INTEGER PRIMARY KEY AUTOINCREMENT,
	ParcelTypeId		INT,		-- where did this message come from
	MesasgeText			TEXT,		-- the raw received message
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
	MesasgeText			TEXT,		-- what will be sent back to user
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

COMMIT;