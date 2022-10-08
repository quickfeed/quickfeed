import { Timestamp } from "google-protobuf/google/protobuf/timestamp_pb"
import { oldFormattedTime, getFormattedTime } from "../Helpers"

describe('Timestamppb', () => {
    it('test timestamp formatting', () => {
        const tsLayout = "2006-01-02T15:04:05"
        const str = oldFormattedTime(tsLayout)

        const timestamp = Timestamp.fromDate(new Date(tsLayout))
        const tsObject = timestamp.toObject()
        const str2 = getFormattedTime(tsObject)
        expect(str2).toBe(str)
    })
})
