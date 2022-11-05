package Logger

import (
	"fmt"
	"os"
	"time"
)

const (
	DEBUG_LOGS     = "DEBUG_LOGS"
	ESSENTIAL_LOGS = "ESSENTIAL_LOGS"
)

var ( 
	DebugLogs int16
	EssentialLogs int16
)

const (
	MASTER = iota
	WORKER
	LOCK_SERVER
	CLUSTER
	CRAWLING
	DATABASE
	MESSAGE_Q
	SERVER
	CACHE
)

const (
	LOG_INFO = iota
	LOG_ERROR
	LOG_DELAY
	LOG_DEBUG
	LOG_TASK_DONE
	LOG_JOB_DONE
	LOG_MILESTONE
	LOG_REQUEST
)

const (
	ESSENTIAL = 1
	DEBUGGING = 2
)

func print(essential int, format string, a ...interface{}) {
	if DebugLogs == 1 || (EssentialLogs == 1 && essential == ESSENTIAL) {
		fmt.Printf(format, a...)
	}
}

func FailOnError(role int, essential int, format string, a ...interface{}) {
	format = beautifyLogs(role, essential, format, LOG_ERROR)
	print(essential, format, a...)
	os.Exit(1)
}

func LogInfo(role int, essential int, format string, a ...interface{}) {
	format = beautifyLogs(role, essential, format, LOG_INFO)
	print(essential, format, a...)
}

func LogError(role int, essential int, format string, a ...interface{}) {
	format = beautifyLogs(role, essential, format, LOG_ERROR)
	print(essential, format, a...)
}

func LogDelay(role int, essential int, format string, a ...interface{}) {
	format = beautifyLogs(role, essential, format, LOG_DELAY)
	print(essential, format, a...)
}

func LogDebug(role int, essential int, format string, a ...interface{}) {
	format = beautifyLogs(role, essential, format, LOG_DEBUG)
	print(essential, format, a...)
}

func LogTaskDone(role int, essential int, format string, a ...interface{}) {
	format = beautifyLogs(role, essential, format, LOG_TASK_DONE)
	print(essential, format, a...)
}

func LogJobDone(role int, essential int, format string, a ...interface{}) {
	format = beautifyLogs(role, essential, format, LOG_JOB_DONE)
	print(essential, format, a...)
}

func LogMilestone(role int, essential int, format string, a ...interface{}) {
	format = beautifyLogs(role, essential, format, LOG_MILESTONE)
	print(essential, format, a...)
}

func LogRequest(role int, essential int, format string, a ...interface{}) {
	format = beautifyLogs(role, essential, format, LOG_REQUEST)
	print(essential, format, a...)
}

func beautifyLogs(role int, essential int, format string, logType int) string {
	additionalInfo := determineRole(role)

	switch logType {
	case LOG_INFO:
		additionalInfo = Green + additionalInfo + "INFO: "
	case LOG_ERROR:
		additionalInfo = Red + additionalInfo + "ERROR: "
	case LOG_DELAY:
		additionalInfo = Yellow + additionalInfo + "DELAY: "
	case LOG_DEBUG:
		additionalInfo = Purple + additionalInfo + "DEBUG: "
	case LOG_TASK_DONE:
		additionalInfo = Cyan + additionalInfo + "TASK_DONE: "
	case LOG_JOB_DONE:
		additionalInfo = White + additionalInfo + "JOB_DONE: "
	case LOG_MILESTONE:
		additionalInfo = Blue + additionalInfo + "MILESTONE: "
	case LOG_REQUEST:
		additionalInfo = Green + additionalInfo + "MILESTONE: "
	default:
		additionalInfo = Blue + additionalInfo + "DEFAULT: "
	}

	additionalInfo += makeTimestamp() + " -> "

	if format[len(format)-1] != '\n' {
		format += "\n"
	}
	format += Reset + "\n" //reset the terminal color

	return additionalInfo + format
}

func makeTimestamp() string {
	return time.Now().Format("01-02-2006 15:04:05")
}

func determineRole(role int) string {

	switch role {
	case MASTER:
		return "MASTER-> "
	case WORKER:
		return "WORKER-> "
	case LOCK_SERVER:
		return "LOCK_SERVER-> "
	case CLUSTER:
		return "CLUSTER-> "
	case CRAWLING:
		return "CRAWLING-> "
	case MESSAGE_Q:
		return "MESSAGE_Q-> "
	case DATABASE:
		return "DATABASE-> "
	case SERVER:
		return "SERVER-> "
	case CACHE:
		return "CACHE-> "
	default:
		return "UNKNOWN-> "
	}
}
