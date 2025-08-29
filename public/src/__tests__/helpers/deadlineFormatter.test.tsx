import { deadlineFormatter, getFormattedTime, TableColor } from "../../Helpers"
import { timeStamp } from "../TestHelpers"


describe("DeadlineFormatter", () => {
    const today = timeStamp({ hours: 1 })
    const expectedDeadlineTextToday = `${59 - (new Date()).getMinutes()} minutes to deadline!`

    const twoMonthsAgo = timeStamp({ months: -1 })
    const fourDaysUntilDeadline = timeStamp({ days: 4 })
    const fourDaysAgo = timeStamp({ days: -4 })
    const twoDaysUntilDeadline = timeStamp({ days: 2 })

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
            deadline: twoMonthsAgo, scoreLimit, submissionScore: scoreLimit,
            deadlineInfo: { className: TableColor.GREEN, message: "Expired 31 days ago", time: getFormattedTime(twoMonthsAgo, true) }
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
    ]

    test.each(tests)("Expected className: $deadlineInfo.className and message: $deadlineInfo.message", ({ deadline, scoreLimit, submissionScore, deadlineInfo }) => {
        const result = deadlineFormatter(deadline, scoreLimit, submissionScore)
        expect(result).toStrictEqual(deadlineInfo)
    })
})
