import CourseCreationInfo from "../components/admin/CourseCreationInfo"
import {configure, shallow} from "enzyme"
import Adapter from "@wojtekmaj/enzyme-adapter-react-17"
import React from "react"

configure({ adapter: new Adapter() });

const title = "Create Course"
let wrapped = shallow(<CourseCreationInfo></CourseCreationInfo>);
describe("Title should be equal to", () => {
    it('Should render correctly', ()=> {
        expect(wrapped).toMatchSnapshot();
    });

    it('Renders titles children', () => {
        expect(wrapped.find('h1').text()).toEqual(title)
    });
});
