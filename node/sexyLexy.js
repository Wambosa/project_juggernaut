/*Created by admin on 9/1/2015.*/

var moment = require('moment');
var sqlite = require('sqlite3').verbose();

var actionMap = []; //no need for this after debug.

main();

function main(){

    var db = new sqlite.Database('someday.db');

    db.all('SELECT * FROM MindVocabulary', [], function(vocabErr, vocabRows){

        if(vocabErr){
            process.exit(vocabErr);}

        var cachedReasoning = [];
        var callbackCount = 0;

        vocabRows.forEach(function(vocab){

            var query =
                ' \
                SELECT * \
                FROM ParameterReasoningGrouping map \
                JOIN ParameterReasoning pr \
                ON pr.ParameterReasoningId = map.ParameterReasoningId \
                WHERE map.MindVocabularyId = $mindVocabularyId \
                ';

            db.all(query, {$mindVocabularyId : vocab.MindVocabularyId}, function(reasoningErr, reasoningRows){

                if(reasoningErr){
                    process.exit(reasoningErr);}

                cachedReasoning[vocab.MindVocabularyId] = vocab;
                cachedReasoning[vocab.MindVocabularyId].parameterReasonings = reasoningRows;
                callbackCount++;
                if(callbackCount === vocabRows.length){
                    ProcessNewMessages(db, cachedReasoning);}
                });
        });
    });

    db.close();
}

function ProcessNewMessages(db, cachedReasoning){

    db.each('SELECT * FROM ReceivedMessage WHERE ParseStatusId != 3', function(err, row){

        var potentialActions = GetPotentialActions(cachedReasoning, row.MessageText);
        var bestAction = {totalCertainty:0};

        potentialActions.forEach(function(action) {

            if (action.totalCertainty > bestAction.totalCertainty) {
                bestAction = action;}
        });

        console.log(bestAction);
        //todo: just take the highest action certainty and create a job record with it, or try to extract the arguments
        //todo: using the argument formula table. {0}\s.* where the variable gets replaces with the original regex
    });
}

function GetPotentialActions(cachedReasoning, text){
    var potentialActions = [];

    cachedReasoning.forEach(function(triggerPhrase){
        if(text.match(triggerPhrase.Regex)) {

            if (potentialActions[triggerPhrase.MindVocabularyId]) {
                potentialActions[triggerPhrase.MindVocabularyId].totalCertainty += triggerPhrase.Certainty;
            } else {
                potentialActions[triggerPhrase.MindVocabularyId] = {
                    humanMessage: text,
                    simpleWord: triggerPhrase.SimpleWord,
                    totalCertainty: triggerPhrase.Certainty,
                    parameters: []
                };
            }

            triggerPhrase.parameterReasonings.forEach(function (paramReasoning) {
                var magicString = paramReasoning.Regex.replace('{0}', triggerPhrase.Regex); //this should not fire unless the param search contains the replaceable syntax. only supports one replace right now.
                var tryArg = text.match(new RegExp(magicString, 'i'));

                if (tryArg) {

                    potentialActions[triggerPhrase.MindVocabularyId].totalCertainty += 100; // if args are discovered, then this is more likely to be the best option

                    if(paramReasoning.IsIncludeOriginRegex) {
                        potentialActions[triggerPhrase.MindVocabularyId].parameters[paramReasoning.ParameterName] = tryArg[0].replace(new RegExp(triggerPhrase.Regex), '').trim();
                    } else {
                        potentialActions[triggerPhrase.MindVocabularyId].parameters[paramReasoning.ParameterName] = tryArg[0].trim();
                    }
                }
            });
        }
    });

    return potentialActions;
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