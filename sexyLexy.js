/*Created by admin on 9/1/2015.*/

var moment = require('moment');
var sqlite = require('sqlite3').verbose();

var db;
var verbosity = 10;
var batchAmount = 100;

main();

//todo: take arg for both database names and verbosity level (also add batch amount)
// SexyLexy -v 10 -b 25 -lex 'bin/sexylexy.db3' -db 'bin/someday.db3'
function main(){

    //todo: set the global db here and the lex db here
    db = new sqlite.Database('bin/someday.db3', function(err){

        if(err){
            process.exit(err);}

        GetCoreLogic(function(actionIds, cachedReasoning){
            ProcessNewMessages(actionIds, cachedReasoning);});

        db.close();
    });
}

function GetCoreLogic(callback){

    db.all('SELECT * FROM Action', [], function(actionErr, actionRows){

        if(actionErr){
            process.exit(actionErr);}

        var actionIds = {};

        actionRows.forEach(function(row){
            actionIds[row.ActionName] = row.ActionId;});

        db.all('SELECT * FROM TriggerPhrase', [], function(vocabErr, vocabRows){

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
                    WHERE map.TriggerPhraseId = $triggerPhraseId \
                    ';

                db.all(query, {$triggerPhraseId : vocab.TriggerPhraseId}, function(reasoningErr, reasoningRows){

                    if(reasoningErr){
                        process.exit(reasoningErr);}

                    cachedReasoning[vocab.TriggerPhraseId] = vocab;
                    cachedReasoning[vocab.TriggerPhraseId].parameterReasonings = reasoningRows;
                    callbackCount++;
                    if(callbackCount === vocabRows.length){
                        callback(actionIds, cachedReasoning);}
                });
            });
        });
    });
}

function ProcessNewMessages(actionIds, cachedReasoning){

    var messageQuery = '\
        SELECT * \
        FROM ReceivedMessage \
        WHERE ParseStatusId = 1 \
        ORDER BY ReceivedMessageId ASC \
        LIMIT ' + batchAmount;

    db.each(messageQuery, function(err, message){

        if(err){
            process.exit(err);}

        var potentialActions = GetPotentialActions(cachedReasoning, message.MessageText);
        var bestAction = {totalCertainty:0};

        potentialActions.forEach(function(action) {

            if (action.totalCertainty > bestAction.totalCertainty) {
                bestAction = action;}
        });

        console.log(bestAction);//todo: verbose only

        if(!bestAction["actionName"]){

            message.ParseStatusId = 4; /* 4 means failed to parse */

            UpdateMessage(
                message,
                function(err){if(err){process.exit(err);}}
            );
            return;
        }

        if(!actionIds[bestAction.actionName]){
            process.exit("Someday Database does not have an action named "+bestAction.actionName);}

        CreateJob(
            actionIds[bestAction.actionName],
            bestAction.totalCertainty,
            message.UserId,
            function(err){

                if(err){
                    process.exit(err);}

                CreateJobMetadata(this.lastID, bestAction.parameters);

                message.JobId = this.lastID;
                message.ParseStatusId = 3; /* 3 means success */

                UpdateMessage(
                    message,
                    function(err){if(err){process.exit(err);}}
                );
            }
        );
    });
}

function GetPotentialActions(cachedReasoning, text){
    var potentialActions = [];

    cachedReasoning.forEach(function(triggerPhrase){
        if(text.match(triggerPhrase.Regex)) {

            potentialActions[triggerPhrase.TriggerPhraseId] = {
                humanMessage: text,
                actionName: triggerPhrase.ActionName,
                totalCertainty: triggerPhrase.Certainty,
                parameters: {}
            };

            triggerPhrase.parameterReasonings.forEach(function (paramReasoning) {
                var magicString = paramReasoning.Regex.replace('{0}', triggerPhrase.Regex); //this should not fire unless the param search contains the replaceable syntax. only supports one replace right now.
                var tryArg = text.match(new RegExp(magicString, 'i'));

                if (tryArg) {

                    if(tryArg != triggerPhrase.Regex) { //todo: test this. may need to strip of special charas before this will work
                        // if args are discovered, then this is more likely to be the best option
                        potentialActions[triggerPhrase.TriggerPhraseId].totalCertainty += 100;
                    }
                    //(this IF may not be needed if i move sexyLexy out of the js regex limitations)
                    if(paramReasoning.IsIncludeOriginRegex) {
                        potentialActions[triggerPhrase.TriggerPhraseId].parameters[paramReasoning.ParameterName] = tryArg[0].replace(new RegExp(triggerPhrase.Regex), '').trim();
                    } else {
                        potentialActions[triggerPhrase.TriggerPhraseId].parameters[paramReasoning.ParameterName] = tryArg[0].trim();
                    }
                }
            });
        }
    });

    return potentialActions;
}

function CreateJob(actionId, certainty, userId, callback){

    var now = moment().utc().format('YYYY-MM-DD HH:mm:ss.SSS');

    var query =
        'INSERT INTO Job \
        (JobStatusId, ActionId, Certainty, UserId, CreatedOn, LastUpdated) \
        VALUES \
        (1, $actionId, $certainty, $userId, $createdOn, $lastUpdated)';

    db.run(query, {
        $actionId : actionId,
        $certainty : certainty,
        $userId : userId,
        $createdOn : now,
        $lastUpdated : now
    }, callback);
}

function CreateJobMetadata(jobId, parameters){

    var query =
        'INSERT INTO JobMetadata \
        (JobId, Key, Value) \
        VALUES \
        ($jobId, $key, $value)';

    Object.getOwnPropertyNames(parameters).forEach(function(paramKey){

        db.run(query, {
            $jobId : jobId,
            $key : paramKey,
            $value : parameters[paramKey]
        },function(err){
            if(err){process.exit(err);}}
        );
    });
}

function UpdateMessage(mutatedMessageRow, callback){

    var now = moment().utc().format('YYYY-MM-DD HH:mm:ss.SSS');

    var query =
        'UPDATE ReceivedMessage \
        SET \
        JobId = $jobId, \
        ParseStatusId = $parseStatusId, \
        LastUpdated = $lastUpdated \
        WHERE ReceivedMessageId = $receivedMessageId';

    db.run(query, {
        $receivedMessageId : mutatedMessageRow.ReceivedMessageId,
        $jobId : mutatedMessageRow.JobId,
        $parseStatusId : mutatedMessageRow.ParseStatusId,
        $lastUpdated : now
    }, callback);
}