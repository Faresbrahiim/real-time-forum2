console.log("Register script loaded");

import { ws } from "./script.js"
import { fetchPosts } from './posts.js';

// signup home navbar
document.getElementById("homesign").addEventListener("click", function (e) {
    document.getElementById('signUpSection').style.display = "none";
    document.getElementById('logInSection').style.display = "block";
    document.getElementById("homesign").style.display = "none";
})

// switch to signUp form
document.getElementById("signUpLink").addEventListener("click", function (e) {
    const form = document.getElementById('myLogInForm');
    const errorDiv = document.getElementById('errorMessage');
    const successDiv = document.getElementById('successMessage');
    errorDiv.style.display = 'none';
    successDiv.style.display = 'none';
    e.preventDefault();
    document.getElementById("logInSection").style.display = "none";
    document.getElementById("signUpSection").style.display = "block";
    document.getElementById("homesign").style.display = "block";
    console.log("Switched to Sign Up section");

})

// signUp form submit
document.addEventListener('DOMContentLoaded', function () {
    const signUpForm = document.getElementById('mySignUpForm');
    if (signUpForm) {
        signUpForm.addEventListener('submit', handleSignUp);
    }
});

// signUp function
async function handleSignUp(event) {
    event.preventDefault();
    const form = document.getElementById('mySignUpForm');
    const errorDiv = document.getElementById('errorMessage');
    const successDiv = document.getElementById('successMessage');
    // Hide previous messages
    errorDiv.style.display = 'none';
    successDiv.style.display = 'none';

    // Convert form data to URL-encoded string instead of FormData
    const formData = new FormData(form);
    const urlEncodedData = new URLSearchParams(formData).toString();

    try {
        const response = await fetch('/api/signup', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/x-www-form-urlencoded',
            },
            body: urlEncodedData
        });

        const result = await response.json();

        if (result.success) {
            successDiv.textContent = result.message;
            successDiv.style.display = 'block';

            setTimeout(() => {
                document.getElementById("signUpSection").style.display = "none";
                document.getElementById("logInSection").style.display = "block";
                form.reset();
            }, 2000);

        } else {
            errorDiv.textContent = result.error;
            errorDiv.style.display = 'block';
        }

    } catch (error) {
        errorDiv.textContent = 'Network error. Please try again.';
        errorDiv.style.display = 'block';
        console.error('Error:', error);
    }

    return false;
}

// log in form submit
document.addEventListener('DOMContentLoaded', function () {
    const logInForm = document.getElementById('myLogInForm');
    if (logInForm) {
        logInForm.addEventListener('submit', handleLogIn);
    }
});

// log in function
async function handleLogIn(event) {
    event.preventDefault();

    const form = document.getElementById('myLogInForm');
    const errorDiv = document.getElementById('logerrorMessage');
    const successDiv = document.getElementById('logsuccessMessage');

    // Hide previous messages
    errorDiv.style.display = 'none';
    successDiv.style.display = 'none';
    // Convert form data to URL-encoded string instead of FormData
    const formData = new FormData(form);
    const urlEncodedData = new URLSearchParams(formData).toString();

    try {
        const response = await fetch('/api/login', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/x-www-form-urlencoded',
            },
            body: urlEncodedData
        });

        const result = await response.json();

        if (result.success) {
            window.afterLogin(); // call it here 
            successDiv.textContent = result.message;
            successDiv.style.display = 'block';
            setTimeout(() => {
                document.getElementById("logInSection").style.display = "none";
                document.getElementById("logout").style.display = "block";
                document.getElementById("createicon").style.display = "block";
                document.getElementById("home").style.display = "block";
                form.reset();
                fetchPosts();
            }, 2000);
        } else {
            errorDiv.textContent = result.error;
            errorDiv.style.display = 'block';
        }
    } catch (error) {
        errorDiv.textContent = 'Network error. Please try again.';
        errorDiv.style.display = 'block';
        console.error('Error:', error);
    }
    return false;
}

// logOut function
document.getElementById("logout").addEventListener("click", function (e) {
    e.preventDefault();
    fetch('/api/logout', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        }
    })
        .then(response => response.json())
        .then(data => {
            if (data.success) {
                console.log("Logged out successfully");
                document.getElementById("freinds").style.display = "none";

                // document.getElementById("userList").style.display="none";

                localStorage.setItem('session_logged_out', Date.now());
                localStorage.setItem('logged_out', 'true');
                localStorage.removeItem('logged_out');

                document.getElementById("logInSection").style.display = "block";
                document.getElementById("signUpSection").style.display = "none";
                document.getElementById("logout").style.display = "none";
                document.getElementById("home").style.display = "none";
                document.getElementById("createicon").style.display = "none";
                document.getElementById("feedPost").style.display = "none";
                document.getElementById('logsuccessMessage').style.display = 'none';
                document.getElementById('fullSinglePost').style.display = 'none';
                document.getElementById('chatBox').style.display = 'none';
                document.getElementById("createPost").style.display = 'none'
                ws.close()

            } else {
                console.error("Logout failed:", data.error);
            }
        })
        .catch(error => console.error('Error during logout:', error));
})