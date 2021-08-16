import React from "react"
import {Bar} from 'react-chartjs-2'
import { useParams } from "react-router-dom";
import { Submission } from "../../proto/ag/ag_pb";
import { useAppState } from "../overmind";

const StatisticsView = () => {
    const course = useParams<{id?: string}>()
    const courseID = Number(course.id)
    let d: number[] = [0, 0]
    const state = useAppState()
    const extractData = () => {
        state.courseSubmissions[courseID]?.forEach(link => {
            link.submissions?.forEach(s => {
                if (!s.getSubmission()) {
                    return
                }
                let c = s.getSubmission()?.getStatus()
                if (c === Submission.Status.APPROVED) {
                    d[0] += 1
                }
                else {
                    d[1] += 1
                }
            })
        })
    }
    extractData()
    let data = {
        labels: ['Approved', 'Not Approved'],
        datasets: [
          {
            label: '# Approved',
            data: d,
            backgroundColor: [
              'rgba(255, 99, 132, 0.2)',
              'rgba(54, 162, 235, 0.2)',
            ],
            borderColor: [
              'rgba(255, 99, 132, 1)',
              'rgba(54, 162, 235, 1)',
            ],
            borderWidth: 1,
          },
        ],
    };

    return (
        <div>
            
        </div>
    )
}

export default StatisticsView