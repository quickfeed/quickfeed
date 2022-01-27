import { derived } from "overmind"
import { Enrollment, Group } from "../../../../proto/ag/ag_pb"


type State = {
    group: Group
    name: string
    users: number[]
    enrollments: Enrollment[]
    edit: boolean
}

export const state: State = {
    group: new Group,
    name: "",
    users: [],
    enrollments: [],
    edit: false
}
