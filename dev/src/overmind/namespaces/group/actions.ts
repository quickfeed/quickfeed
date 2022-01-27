import { json } from 'overmind'
import { Context } from '../../'
import { Group, User } from '../../../../proto/ag/ag_pb'
import { Color } from '../../../Helpers'
import { success } from '../../actions'

export const resetGroup = ({ state }: Context) => {
    state.group.group = new Group
    state.group.name = ""
    state.group.users = []
    state.group.edit = false
}

export const setName = ({ state }: Context, name: string) => {
    state.group.name = name
}

export const addUsers = ({ state }: Context, user: User) => {
    state.group.users.push(user.getId())
}

export const createGroup = async ({ state, actions, effects }: Context): Promise<boolean | undefined> => {
    const name = state.group.name
    if (isValid(name, state.group.users)) {
        // false if name is not alphanumerical
        state.group.group.setUsersList(state.group.users.map(user => new User().setId(user)))
        state.group.group.setCourseid(state.activeCourse)
        state.group.group.setName(name)
        const success = await effects.grpcMan.createGroup(json(state.group.group))
        if (success.data) {
            state.enrollmentsByCourseID[state.activeCourse].setGroup(success.data)
            state.userGroup[state.activeCourse] = success.data
            actions.group.resetGroup()
            return true
        }
    } else {
        actions.alert({ text: "Group name must be alphanumerical and between 1 and 25 characters", color: Color.RED })
    }
}

export const updateUsers = ({ state }: Context, user: User) => {
    if (user.getId() === state.self.getId()) {
        return
    }
    const users = state.group.users
    const index = indexOf(users, user)
    if (index >= 0) {
        // Remove user with id from group
        users.splice(index, 1)
    } else {
        users.push(user.getId())
    }
}

export const setGroup = ({ state }: Context, group: Group) => {
    Object.assign(state.group.group, json(group))
    state.group.name = group.getName()
    for (const user of group.getUsersList()) {
        state.group.users.push(user.getId())
    }
}

export const updateGroup = async ({ state, actions, effects }: Context): Promise<boolean> => {
    const name = state.group.name
    if (!isValid(name, state.group.users)) {
        actions.alert({ text: "Group name must be alphanumerical and group must have at least one member", color: Color.RED })
        return false
    }
    const grp = new Group()
    grp.setId(state.group.group.getId())
    grp.setUsersList(state.group.users.map(user => new User().setId(user)))
    grp.setCourseid(state.activeCourse)
    grp.setName(name)

    const response = await effects.grpcMan.updateGroup(grp)
    if (success(response) && response.data) {
        Object.assign(state.groups[state.activeCourse].find(group => group.getId() === state.group.group.getId())!, response.data)
        actions.group.resetGroup()
        return true
    } else {
        actions.alertHandler(response)
        return false
    }
}

const indexOf = (users: number[], user: User): number => {
    for (let i = 0; i < users.length; i++) {
        if (users[i] === user.getId()) {
            return i
        }
    }
    return -1
}

// isValid checks if the group name is alphanumerical and between 1 and 25 characters, and if the group has at least one member
const isValid = (name: string, users: number[]): boolean => {
    if (name.match(/^[a-zA-Z0-9]+$/i) && name.length > 0 && name.length <= 25 && users.length > 0) {
        return true
    }
    return false
}
