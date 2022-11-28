jobContent = './jobContent.txt'

def distribute(jobContent):
    f = open(jobContent, 'r')

    tasks = []

    file_contents = f.readlines()

    for taskString in file_contents:
        tasks.append(taskString.replace('\n', ''))

    f.close()

    return tasks