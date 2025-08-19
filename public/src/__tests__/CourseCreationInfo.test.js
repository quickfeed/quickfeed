import CourseCreationInfo from "../components/admin/CourseCreationInfo";
import React from "react";
import { render } from "@testing-library/react";
describe("Title should be equal to", () => {
    it('Renders titles children', () => {
        const title = "Create Course";
        const wrapped = render(React.createElement(CourseCreationInfo, null));
        const element = wrapped.getByRole("heading");
        expect(element.textContent).toEqual(title);
    });
});
