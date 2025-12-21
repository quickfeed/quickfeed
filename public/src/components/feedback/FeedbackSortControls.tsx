import React from 'react'

interface FeedbackSortControlsProps {
    sortOrder: 'asc' | 'desc' | 'none'
    setSortOrder: (order: 'asc' | 'desc' | 'none') => void
    feedbackCount: number
}

export const FeedbackSortControls: React.FC<FeedbackSortControlsProps> = ({
    sortOrder,
    setSortOrder,
    feedbackCount
}) => {
    const toggleSort = () => {
        if (sortOrder === 'none') setSortOrder('asc')
        else if (sortOrder === 'asc') setSortOrder('desc')
        else setSortOrder('none')
    }

    const getSortIcon = () => {
        if (sortOrder === 'asc') return 'fa-sort-amount-asc'
        if (sortOrder === 'desc') return 'fa-sort-amount-desc'
        return 'fa-sort'
    }

    const getSortLabel = () => {
        if (sortOrder === 'asc') return 'Time (Low to High)'
        if (sortOrder === 'desc') return 'Time (High to Low)'
        return 'Sort by Time'
    }

    return (
        <div className="d-flex justify-content-between align-items-center mb-3">
            <h3 className="mb-0">
                <i className="fa fa-list mr-2" />
                Individual Feedback ({feedbackCount})
            </h3>
            <button
                className={`btn btn-outline-secondary ${sortOrder !== 'none' ? 'active' : ''}`}
                onClick={toggleSort}
            >
                <i className={`fa ${getSortIcon()} mr-2`} />
                {getSortLabel()}
            </button>
        </div>
    )
}

export default FeedbackSortControls
