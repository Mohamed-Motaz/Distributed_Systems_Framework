import json

jobContent = './jobContent.txt'

def distribute(jobContent):
    f = open(jobContent, 'r')

    tasks = []

    file_contents = f.readlines()

    for taskString in file_contents:
        tasks.append(taskString.replace('\n', ''))

    f.close()

    return tasks

tasks = distribute(jobContent)


def process(taskFilePath):
    f = open(taskFilePath, 'r+')

    wordsCount = {}

    file_contents = f.read().split()

    for word in file_contents:
        currentWord = word.lower()
        if( wordsCount.get(currentWord) == None ):
            wordsCount[currentWord] = 1
        else:
            wordsCount[currentWord] = wordsCount.get(currentWord) + 1

    f.seek(0)
    f.truncate(0)
    f.write(json.dumps(wordsCount))
    f.close()


def aggregate(jobContent):
    f = open(jobContent, 'r')

    tasksFilePaths = []

    file_contents = f.readlines()

    for taskString in file_contents:
        tasksFilePaths.append(taskString.replace('\n', ''))

    f.close()

    totalWordsCount = {}

    i = 0
    while i < len(tasksFilePaths):
        f = open(tasksFilePaths[i], 'r')

        wordsDict = f.read()
        wordsDict = json.loads(wordsDict)

        if( i == 0 ):
            totalWordsCount = wordsDict
        else:
            for word in wordsDict:
                currentWord = word.lower()
                if (totalWordsCount.get(currentWord) is None):
                    totalWordsCount[currentWord] = wordsDict.get(currentWord)
                else:
                    totalWordsCount[currentWord] = wordsDict.get(currentWord) + totalWordsCount.get(currentWord)

        f.close()

        i += 1

    return totalWordsCount

print(aggregate(jobContent))