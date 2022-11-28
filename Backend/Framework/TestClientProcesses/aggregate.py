import json


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

        if i == 0:
            totalWordsCount = wordsDict
        else:
            for word in wordsDict:
                currentWord = word.lower()
                if totalWordsCount.get(currentWord) is None:
                    totalWordsCount[currentWord] = wordsDict.get(currentWord)
                else:
                    totalWordsCount[currentWord] = wordsDict.get(currentWord) + totalWordsCount.get(currentWord)

        f.close()

        i += 1

    return totalWordsCount
