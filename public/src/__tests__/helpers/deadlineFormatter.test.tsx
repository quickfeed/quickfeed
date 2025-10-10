import { deadlineFormatter, getFormattedTime, TableColor } from "../../Helpers"
import { timeStamp } from "../TestHelpers"


describe("DeadlineFormatter", () => {
    const today = timeStamp({ hours: 1 })
    const expectedDeadlineTextToday = `${59 - (new Date()).getMinutes()} minutes to deadline!`

    const fourDaysUntilDeadline = timeStamp({ days: 4 })
    const twoDaysUntilDeadline = timeStamp({ days: 2 })
    const fourDaysAgo = timeStamp({ days: -4 })
    const twentyEightDaysAgo = timeStamp({ days: -28 })
    const thirtyDaysAgo = timeStamp({ days: -30 })
    const thirtyOneDaysAgo = timeStamp({ days: -31 })
    const thirtyTwoDaysAgo = timeStamp({ days: -32 })
    const fiftyDaysAgo = timeStamp({ days: -50 })
    const sixtyDaysAgo = timeStamp({ days: -60 })

    const scoreLimit = 50
    const tests = [
        {
            deadline: today, scoreLimit, submissionScore: 0,
            deadlineInfo: { className: TableColor.RED, message: expectedDeadlineTextToday, time: getFormattedTime(today, true) }
        },
        {
            deadline: today, scoreLimit, submissionScore: scoreLimit,
            deadlineInfo: { className: TableColor.GREEN, message: expectedDeadlineTextToday, time: getFormattedTime(today, true) }
        },
        {
            deadline: fourDaysAgo, scoreLimit, submissionScore: 0,
            deadlineInfo: { className: TableColor.RED, message: "Expired 4 days ago", time: getFormattedTime(fourDaysAgo, true) }
        },
        {
            deadline: twoDaysUntilDeadline, scoreLimit, submissionScore: 0,
            deadlineInfo: { className: TableColor.ORANGE, message: "2 days to deadline!", time: getFormattedTime(twoDaysUntilDeadline, true) }
        },
        {
            deadline: fourDaysUntilDeadline, scoreLimit, submissionScore: 0,
            deadlineInfo: { className: TableColor.BLUE, message: "4 days to deadline", time: getFormattedTime(fourDaysUntilDeadline, true) }
        },
        {
            deadline: twentyEightDaysAgo, scoreLimit, submissionScore: scoreLimit,
            deadlineInfo: { className: TableColor.GREEN, message: "Expired 28 days ago", time: getFormattedTime(twentyEightDaysAgo, true) }
        },
        {
            deadline: thirtyDaysAgo, scoreLimit, submissionScore: scoreLimit,
            deadlineInfo: { className: TableColor.GREEN, message: "Expired 30 days ago", time: getFormattedTime(thirtyDaysAgo, true) }
        },
        {
            deadline: thirtyOneDaysAgo, scoreLimit, submissionScore: scoreLimit,
            deadlineInfo: { className: TableColor.GREEN, message: "Expired 31 days ago", time: getFormattedTime(thirtyOneDaysAgo, true) }
        },
        {
            deadline: thirtyTwoDaysAgo, scoreLimit, submissionScore: scoreLimit,
            deadlineInfo: { className: TableColor.GREEN, message: "Expired 32 days ago", time: getFormattedTime(thirtyTwoDaysAgo, true) }
        },
        {
            deadline: fiftyDaysAgo, scoreLimit, submissionScore: scoreLimit,
            deadlineInfo: { className: TableColor.GREEN, message: "Expired 50 days ago", time: getFormattedTime(fiftyDaysAgo, true) }
        },
        {
            deadline: sixtyDaysAgo, scoreLimit, submissionScore: scoreLimit,
            deadlineInfo: { className: TableColor.GREEN, message: "Expired 60 days ago", time: getFormattedTime(sixtyDaysAgo, true) }
        },
    ]

    test.each(tests)("Expected className: $deadlineInfo.className and message: $deadlineInfo.message", ({ deadline, scoreLimit, submissionScore, deadlineInfo }) => {
        const result = deadlineFormatter(deadline, scoreLimit, submissionScore)
        expect(result).toStrictEqual(deadlineInfo)
    })
})
