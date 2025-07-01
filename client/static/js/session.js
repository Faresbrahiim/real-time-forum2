console.log("Session script loaded");

import { fetchPosts } from './posts.js';

document.addEventListener('DOMContentLoaded', checkSession);

window.addEventListener('storage', (event) => {
    // Make sure the key is `session_logged_out`
    if (event.key === 'session_logged_out') {
        // Handle session logout in this tab
        checkSession();
    }
});

async function checkSession() {
    console.log("Checking session status...");
    try {
        
        const response = await fetch('/api/checksession');
        const result = await response.json();

        if (result.loggedIn) {
            document.getElementById("logInSection").style.display = "none";
            document.getElementById("feedPost").style.display = "grid";
            document.getElementById("logout").style.display = "block";
            document.getElementById("createicon").style.display = "block";
            document.getElementById("home").style.display = "block";
            console.log("User is logged in:", result.username);
            fetchPosts();
        } else {
            document.getElementById("logInSection").style.display = "block";
            document.getElementById("feedPost").style.display = "none";
            document.getElementById("logout").style.display = "none";
            document.getElementById("createicon").style.display = "none";
            document.getElementById("home").style.display = "none";
            document.getElementById("logsuccessMessage").style.display = "none";
            document.getElementById("logerrorMessage").style.display = "none";
            document.getElementById("fullSinglePost").style.display = "none";
            document.getElementById("createPost").style.display = "none";

        }
    } catch (error) {
        console.error('Failed to check session:', error);
    }
}

document.getElementById("home").addEventListener("click", function(e) {
    e.preventDefault();
    document.getElementById("feedPost").style.display = "grid";
    document.getElementById("fullSinglePost").style.display = "none";
    document.getElementById("createPost").style.display = "none";
    document.getElementById("myCreateForm").reset();
    document.getElementById("createErrorMessage").style.display = "none";
    document.getElementById("createSuccessMessage").style.display = "none";
    fetchPosts();
});