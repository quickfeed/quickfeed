import { render } from "@testing-library/react"
import CourseCreationInfo from "../components/admin/CourseCreationInfo"


describe("Title should be equal to", () => {

    it('Renders titles children', () => {
        const title = "Create Course"
        const wrapped = render(<CourseCreationInfo />)
        const element = wrapped.getByRole("heading")
        expect(element.textContent).toEqual(title)
    })
})
