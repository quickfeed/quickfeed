import {updateAdmin}  from "../overmind/actions"
import { User} from "../../proto/ag/ag_pb"
import { createOvermindMock } from "overmind"
import { config } from "../overmind"


describe("Correct permission status should be set", () => {
    it('If user is not admin, promote to admin', () =>{
        const user = new User().setId(1).setName("Test User").setIsadmin(false);
        const mockedOvermind = createOvermindMock(config, (state) => {
            state.self = user
        })
        window.confirm = jest.fn(() => true)
        updateAdmin(mockedOvermind, user)
        var bool = user.getIsadmin()
        expect(bool).toBe(true)
    })
    
    it('If user is admin, demote user', () => {
        const user2 = new User().setId(2).setName("Test User2").setIsadmin(true);
        const mockedOvermind2 = createOvermindMock(config, (state) => {
            state.self = user2
        })
        window.confirm = jest.fn(() => true)
        updateAdmin(mockedOvermind2, user2)
        expect(user2.getIsadmin()).toBe(false)
    })
    
    it('If user does not confirm, dont make any changes', ()=>{
        const user3 = new User().setId(3).setName("Test User3").setIsadmin(true);
        const mockedOvermind2 = createOvermindMock(config, (state) => {
            state.self = user3
        })
        window.confirm = jest.fn(() => false)
        updateAdmin(mockedOvermind2, user3)
        expect(user3.getIsadmin()).toBe(true)
    })
});
