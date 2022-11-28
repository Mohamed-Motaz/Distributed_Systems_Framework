import json

filePath = "./aggregate.txt"

def aggregate():
    f = open(filePath, 'r+')
    tasksFilePaths = []
    file_contents = f.readlines()

    for taskString in file_contents:
        tasksFilePaths.append(taskString.replace('\n', ''))


    totalWordsCount = {}

    i = 0
    while i < len(tasksFilePaths):
        taskF = open(tasksFilePaths[i], 'r')

        wordsDict = taskF.read()
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

        taskF.close()
        i += 1

    f.seek(0)
    f.truncate(0)
    f.write(str(totalWordsCount))
    f.close()

