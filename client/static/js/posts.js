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

    const formData = new FormData(form);
        const selectedCategories = Array.from(document.querySelectorAll('input[name="categories"]:checked'))
                                    .map(input => input.value);

    selectedCategories.forEach(cat => formData.append('categories[]', cat));

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
        document.getElementById("freinds").style.display="flex";
        if (!Array.isArray(posts)) {
            return;
        }
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
        single.innerHTML = '';

        // Create elements with safe text insertion
        const titleEl = document.createElement('h2');
        titleEl.className = 'postTitle';
        titleEl.textContent = postData.title;

        const categoryEl = document.createElement('p');
        categoryEl.className = 'postCategory';
        categoryEl.textContent = 'Category: ' + postData.category;

        const contentEl = document.createElement('p');
        contentEl.className = 'postContent';
        contentEl.textContent = postData.content;

        single.appendChild(titleEl);
        single.appendChild(categoryEl);
        single.appendChild(contentEl);

        Array.from(document.getElementsByClassName('postContent')).forEach(el => {
        el.style.display = 'block'});
        document.getElementById('commentForm').style.display = 'flex';
        document.getElementById('comment').style.display = 'block';
        document.getElementById('fullSinglePost').style.display = 'flex';
    } catch (error) {
        console.error('Error displaying single post:', error);
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

async function comments() {
    try {
        const response = await fetch(`/api/comments/${postid}`);
        if (!response.ok) throw new Error('Failed to fetch comments');
        const commentsData = await response.json();

        const commentsDiv = document.getElementById('comment');
        commentsDiv.style.display = 'block';
        
        // Clear previous comments safely
        commentsDiv.innerHTML = '';

        if (Array.isArray(commentsData) && commentsData.length > 0) {
            commentsData.forEach(comment => {
                const commentDiv = document.createElement('div');
                commentDiv.className = 'comment';

                const contentP = document.createElement('p');
                contentP.className = 'commentContent';
                contentP.textContent = comment.content;

                const authorP = document.createElement('p');
                authorP.className = 'commentAuthor';
                authorP.textContent = 'By: ' + comment.author;

                commentDiv.appendChild(contentP);
                commentDiv.appendChild(authorP);

                commentsDiv.appendChild(commentDiv);
            });
        } else {
            console.log('No comments found for this post.');
            commentsDiv.style.display = 'none';
        }
    } catch (error) {
        console.error('Error loading comments:', error);
    }
}