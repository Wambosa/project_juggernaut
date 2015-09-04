/*Created by admin on 9/1/2015.*/

var moment = require('moment');
var sqlite = require('sqlite3').verbose();
var db = new sqlite.Database('someday.db');

var actionMap = []; //no need for this after debug.

db.each('SELECT * FROM ReceivedMessage WHERE ParseStatusId != 3', function(err, row){
    GetPotentialActions(db, row.MessageText, function(foundActions){

        actionMap.push({
            humanMessage: row.MessageText,
            potentialActions: foundActions
        });
        //todo: just take the highest action certainty and create a job record with it, or try to extract the arguments
        //todo: using the argument formula table. {0}\s.* where the variable gets replaces with the original regex
        console.log({
            humanMessage: row.MessageText,
            potentialActions: foundActions
        });
    });
});

db.close();

function GetPotentialActions(database, text, callback){
    //todo: also get possible args associated with a particular match
    //todo: cache the mind vocabulary on the first run, if it is not null, then use the global cache
    database.all('SELECT * FROM MindVocabulary', [], function(err, rows){

        if(!err){

            var potentialActions = {};

            rows.forEach(function(mindVocabulary){
                if(text.match(mindVocabulary.Regex)){
                    if(potentialActions[mindVocabulary.SimpleWord]){
                        potentialActions[mindVocabulary.SimpleWord] += mindVocabulary.Certainty;
                    }else{
                        potentialActions[mindVocabulary.SimpleWord] = mindVocabulary.Certainty;
                    }

                    //need to create an object that contains certainty later. since i also need named args (which must be obtained via arg formula)
                }
            });

            callback(potentialActions);
        }
    });
}

function CreateJob(database, actionId, certainty, userId, callback){

    var now = moment().utc().format('YYYY-MM-DD HH:mm:ss.SSS');

    var query =
        'INSERT INTO Job \
        (JobStatusId, ActionId, Certainty, UserId, CreatedOn, LastUpdated) \
        VALUES \
        (1, $actionId, $certainty, $userId, $createdOn, lastUpdated)';

    database.run(query, {
        $actionId : actionId,
        $certainty : certainty,
        $userId : userId,
        $createdOn : now,
        $lastUpdated : now
    }, function(err){callback(err);});
}

function UpdateMessage(database, mutatedMessageRow, callback){

    var now = moment().utc().format('YYYY-MM-DD HH:mm:ss.SSS');

    var query =
        'UPDATE ReceivedMessage \
        SET \
        JobId = $jobId, \
        ParseStatusId = $parseStatusId, \
        LastUpdated = $lastUpdated';

    database.run(query, {
        $jobId : mutatedMessageRow.jobId,
        $parseStatusId : mutatedMessageRow.parseStatuId,
        $lastUpdated : now
    }, function(err){callback(err);});
}