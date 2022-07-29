import { Timestamp } from "google-protobuf/google/protobuf/timestamp_pb"
import { getFormattedTime, formattedDate } from "../Helpers"

describe('Timestamppb', () => {
    it('test timestamp formatting', async () => {
        const tsLayout = "2006-01-02T15:04:05"
        const str = getFormattedTime(tsLayout)

        const timestamp = Timestamp.fromDate(new Date(tsLayout))
        const tsObject = timestamp.toObject()
        const str2 = formattedDate(tsObject)
        expect(str2).toBe(str)
    })
})
