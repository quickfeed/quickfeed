import React from 'react'

const ProfileHero = ({ name }: { name: string }) => (
    <div className="hero">
        <div className="hero-content text-center">
            <div className="max-w-2xl">
                <h1 className="text-5xl font-bold mb-4">Hi, {name}</h1>
                <p className="text-lg mb-2">You can edit your user information here.</p>
                <div className="alert alert-warning mt-4">
                    <i className="fa fa-exclamation-triangle" />
                    <span><strong>Use your real name as it appears on Canvas</strong> to ensure that approvals are correctly attributed.</span>
                </div>
            </div>
        </div>
    </div>
)

export default ProfileHero
