let postid = null;

document.addEventListener('DOMContentLoaded', function() {
    console.log('Posts script loaded');
    const postForm = document.getElementById('myCreateForm');
    if (postForm) {
        postForm.addEventListener('submit', handlePostCreation);
    }
});

async function handlePostCreation(event) {
    event.preventDefault();

    const form = document.getElementById('myCreateForm');
    const errorDiv = document.getElementById('createErrorMessage');
    const successDiv = document.getElementById('createSuccessMessage');
    const postsDiv = document.getElementById('feedPost');

    postsDiv.style.display = "none";
    
    // Hide previous messages
    errorDiv.style.display = 'none';
    successDiv.style.display = 'none';

    // Convert form data to URL-encoded string instead of FormData
    const formData = new FormData(form);
    const urlEncodedData = new URLSearchParams(formData).toString();
    
    try {
        const response = await fetch('/api/createpost', {
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
                document.getElementById("createPost").style.display = "none";
                fetchPosts();
                postsDiv.style.display = "grid";
                
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

document.getElementById('createicon').addEventListener('click', function() {
    document.getElementById("createPost").style.display = "block";
    document.getElementById("feedPost").style.display = "none";
    document.getElementById("fullSinglePost").style.display = "none";
    document.getElementById("myCreateForm").reset();
    document.getElementById("createErrorMessage").style.display = "none";
    document.getElementById("createSuccessMessage").style.display = "none";
});

export async function fetchPosts() {
    console.log("Fetching posts...");
    try {
        const response = await fetch('/api/posts');
        if (!response.ok) throw new Error('Failed to fetch posts');
        const posts = await response.json();

        const feedPost = document.getElementById('feedPost');
        feedPost.style.display = 'grid';
        feedPost.innerHTML = '';
        // display freinds section
        document.getElementById("freinds").style.display="block";

        posts.forEach(post => {
            const postDiv = document.createElement('div');
            postDiv.className = 'posts';
            postDiv.id = `${post.id}`;
            postDiv.innerHTML = `
                
                <h2 class="postTitle">${post.title}</h2>
                <p class="postCategory">Category: ${post.category}</p>
                <p class="postContent">${post.content}</p>
            `;

            feedPost.appendChild(postDiv);
        });
    } catch (err) {
        console.error('Error loading posts:', err);
    }
}

document.addEventListener("click", function(e) {
    const post = e.target.closest(".posts");
    if (post) {
        document.getElementById("feedPost").style.display = "none";
        postid = post.id;
        singlePost();
        comments();
    }
});

async function singlePost() {
    try {
        const response = await fetch(`/api/singlepost/${postid}`);
        if (!response.ok) throw new Error('Failed to fetch single post');
        const postData = await response.json();

        const single = document.getElementById('singlePost');
        single.style.display = 'grid';
        single.innerHTML = `
            <h2 class="postTitle">${postData.title}</h2>
            <p class="postCategory">Category: ${postData.category}</p>
            <p class="postContent">${postData.content}</p>
        `;
        Array.from(document.getElementsByClassName('postContent')).forEach(el => {
        el.style.display = 'block'});
        document.getElementById('commentForm').style.display = 'flex';
        document.getElementById('comment').style.display = 'block';
        document.getElementById('fullSinglePost').style.display = 'flex';
    } catch (error) {
        console.error('Error displaying single post:', error);
    }
}

async function comments() {
    try {
        const response = await fetch(`/api/comments/${postid}`);
        if (!response.ok) throw new Error('Failed to fetch comments');
        const commentsData = await response.json();

        const commentsDiv = document.getElementById('comment');
        commentsDiv.style.display = 'block';
        commentsDiv.innerHTML = '';
        if (Array.isArray(commentsData)) {
            commentsData.forEach(comment => {
                const commentDiv = document.createElement('div');
                commentDiv.className = 'comment';
                commentDiv.innerHTML = `
                    <p class="commentContent">${comment.content}</p>
                    <p class="commentAuthor">By: ${comment.author}</p>
                `;
                commentsDiv.appendChild(commentDiv);
            });
        } else {
            console.log('No comments found for this post.');
            document.getElementById('comment').style.display = 'none';
        }
    } catch (error) {
        console.error('Error loading comments:', error);
    }
}

document.addEventListener('DOMContentLoaded', function() {
    const comment = document.getElementById('commentForm');
    if (comment) {
        comment.addEventListener('submit', handleCommentCreation)
    }
});

async function handleCommentCreation(event) {
    event.preventDefault();

    const form = document.getElementById('commentForm');
    const formData = new FormData(form);
    const urlEncodedData = new URLSearchParams(formData).toString();
    const submitBtn = document.getElementById('submitComment');
    const input = document.getElementById('commentFormInput'); 
    const value = input.value.trim();
    const isValid = /[A-Za-z0-9]/.test(value);

    if (!isValid) {
        alert("Comment must contain at least one letter or number");
        return;
    }
    submitBtn.disabled = true;
    submitBtn.value = "Submitting...";
    
    try {
        const response = await fetch(`/api/createcomment/${postid}`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/x-www-form-urlencoded',
            },
            body: urlEncodedData
        });
        const result = await response.json();
            if (result.success) {
                form.reset();
                await comments();
                submitBtn.disabled = false;
                submitBtn.value = "Submit";
        }
    } catch (error) {
        console.error('Error:', error);
        submitBtn.disabled = false;
        submitBtn.value = "Submit";
    }
}