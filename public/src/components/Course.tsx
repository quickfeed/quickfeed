import { RouteProps } from "react-router"
import { Courses } from "../proto/ag_pb"


const Course = (props: RouteProps) => {
    return <h1>Test course, {props.children?.toString()}</h1>
}

export default Course