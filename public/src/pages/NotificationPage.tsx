import React from "react"
import { useActions, useAppState } from "../overmind"
import FormInput from "../components/forms/FormInput"
import { Notification_RecipientType, NotificationSchema, Notification, Enrollment, User, Notification_Type } from "../../proto/qf/types_pb"
import Button, { ButtonType } from "../components/admin/Button"
import { Color, isTeacher } from "../Helpers"
import { create } from "@bufbuild/protobuf"
import Alerts from "../components/alerts/Alerts"
import FormSelect from "../components/forms/FormSelect"

const NotificationPage = () => {
    const state = useAppState()
    const [showForm, setShowForm] = React.useState(false)
    const formButtonText = showForm ? "View Notifications" : "Create Notification"
    const title = showForm ? "Create Notification" : "Notifications"
    const buttonColor = showForm ? Color.BLUE : Color.GREEN

    return (
        <div className="d-flex flex-column align-items-center mt-5">
            <div className="col-4">
                <Alerts />
            </div>
            <div className="d-flex col-4">
                <h1>{title}</h1>
                {state.self.IsAdmin || state.isTeacher ?
                    <div className="ml-auto mt-2">
                        <Button
                            onClick={() => setShowForm(!showForm)}
                            text={formButtonText}
                            color={buttonColor}
                            type={ButtonType.BUTTON}
                        />
                    </div>
                    : null}
            </div>
            {
                showForm
                    ? <NotificationForm />
                    : <NotificationList notifications={state.notifications} />
            }
        </div>
    )
}

const NotificationList = ({ notifications }: { notifications: Notification[] }) => {
    return (
        notifications.map((notification) => (
            <div className="col-sm-4 mb-3" key={notification.ID}>
                <div className="card">
                    <div className="card-body">
                        <h3 className="card-title">
                            {notification.title}
                        </h3>
                        <h5 className="card-text grey">
                            {notification.body}
                        </h5>
                    </div>
                </div>
            </div>
        ))
    )
}


const NotificationForm = () => {
    const state = useAppState()
    const actions = useActions()

    const [notification, setNotification] = React.useState(create(NotificationSchema))
    notification.sender = state.self.ID
    const coursesWithTeacherEnrollment = state.courses.filter(course =>
        state.enrollments.some(enrollment => isTeacher(enrollment) && course.ID === enrollment.courseID)
    )
    const [selectedRecipients, setSelectedRecipients] = React.useState<User[]>([])
    const [selectedCourseID, setSelectedCourse] = React.useState(coursesWithTeacherEnrollment[0].ID)

    // Only admins and teachers can send notifications
    if (!state.self.IsAdmin && !state.isTeacher) {
        return null
    }

    const handleChange = (event: React.FormEvent<HTMLInputElement>) => {
        const { name, value } = event.currentTarget
        switch (name) {
            case "title":
                notification.title = value
                break
            case "body":
                notification.body = value
                break
        }
        setNotification(notification)
    }
    const updateReceivers = (value: Notification_RecipientType) => {
        notification.recipientType = value
        setNotification(notification)
    }

    const recipientTypes = Object.values(Notification_RecipientType).filter(v => typeof v !== 'number')
    if (!state.self.IsAdmin) {
        const index = recipientTypes.indexOf(Notification_RecipientType[Notification_RecipientType.ALL])
        recipientTypes.splice(index, 1)
    }
    const recipientOptions = recipientTypes.map((name, i) => ({
        value: i.toString(),
        key: name,
        text: name,
    }))

    const courseOptions = coursesWithTeacherEnrollment.reverse().map((course) => ({
        value: course.ID.toString(),
        key: course.code + course.year,
        text: course.code + " " + course.year,
    }))

    const notificationTypes = Object.values(Notification_Type).filter(v => typeof v !== 'number')
    const notificationOptions = notificationTypes.map((name, i) => ({
        value: i.toString(),
        key: name,
        text: name,
    }))

    const addRecipient = (userID: bigint) => {
        if (notification.recipientIDs.find((id) => id === userID)) {
            actions.alert({ text: "Recipient is already added", color: Color.RED })
            return
        }
        notification.recipientIDs.push(userID)
        calcRecipients()
    }

    const calcRecipients = () => {
        setNotification(notification)
        setSelectedRecipients(notification.recipientIDs.flatMap((id) => {
            if (state.self.IsAdmin) {
                return state.allUsers.filter((user) => user.ID === id)
            } else {
                return state.courseEnrollments[selectedCourseID.toString()]
                    .map((enrollment) => enrollment.user)
                    .filter((user) => user !== undefined)
                    .filter((user) => user.ID === id)
            }
        }))
    }

    const removeRecipient = (userID: bigint) => {
        const index = notification.recipientIDs.findIndex((id) => id === userID)
        if (index === -1) {
            return
        }
        notification.recipientIDs.splice(index, 1)
        calcRecipients()
    }

    const getCourseData = async () => {
        await actions.getCourseData({ courseID: selectedCourseID })
    }

    if (!state.courseEnrollments[selectedCourseID.toString()]) {
        getCourseData()
    }
    let enrollments: Enrollment[] = []
    if (state.courseEnrollments[selectedCourseID.toString()]) {
        // Clone the enrollments so we can sort them
        enrollments = state.courseEnrollments[selectedCourseID.toString()].slice().filter((enrollment) => {
            return enrollment.user && !notification.recipientIDs.includes(enrollment.user.ID)
        })
    }

    return (
        <>
            <div className="col-sm-4">

                <FormSelect
                    prepend="Course"
                    name="courses"
                    options={notificationOptions}
                    onChange={() => { }}
                />

                <FormInput
                    prepend="Title"
                    name="title"
                    placeholder={"title"}
                    defaultValue={""}
                    onChange={handleChange}
                />
                <FormInput
                    prepend="Body"
                    name="body"
                    placeholder={"description"}
                    defaultValue={""}
                    onChange={handleChange}
                />

                <FormSelect
                    prepend="Recipients Type"
                    name="recipients"
                    options={recipientOptions}
                    onChange={(e) => updateReceivers(Number(e.target.value))}
                />

                <FormSelect
                    prepend="Course"
                    name="courses"
                    options={courseOptions}
                    onChange={(e) => setSelectedCourse(BigInt(e.target.value))}
                />


                <div className="d-flex flex-column">
                    <h2> Recipients </h2>
                    <div className="d-flex flex-wrap m-2">
                        {selectedRecipients.length === 0 ?
                            <h5> No Recipients Selected </h5>
                            : selectedRecipients.map((recipient) => (
                                <button className="m-1 hover-effect" onClick={() => removeRecipient(recipient.ID)} key={recipient.ID}>
                                    {recipient.Name}
                                    <i className="fa fa-times-circle ml-2" />
                                </button>
                            ))
                        }
                    </div>
                </div>


                <Button
                    text="Send"
                    color={Color.GREEN}
                    type={ButtonType.BUTTON}
                    className="float-right mt-2 col-3"
                    onClick={() => actions.sendNotification(notification)}
                />

                {enrollments.length === 0 ?
                    <h2> There are no users enrolled into this course </h2>
                    : (
                        <div className="d-flex flex-wrap mt-5">
                            {enrollments.map((enrollment) => {
                                const user = enrollment.user
                                if (!user) return null

                                return <button className="m-1 hover-effect" onClick={() => addRecipient(user.ID)} key={user.ID}>
                                    {user.Name}
                                    <i className="fa fa-plus ml-2" />
                                </button>
                            })}
                        </div>
                    )
                }
            </div>
        </>
    )

}


export default NotificationPage
