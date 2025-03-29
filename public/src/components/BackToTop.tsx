import React from 'react'

const ScrollToTop = () => {
    window.scrollTo({ top: 0, behavior: "smooth" })
}

const BackToTop: React.FC = () => {
    return (
        <footer className="text-center mt-5">
            <button onClick={ScrollToTop} className="btn align-items-center backToTop">
                <i className="fa fa-arrow-up" />
                <p>Back to top</p>
            </button>
        </footer>
    )
}

export default BackToTop
