document.addEventListener('DOMContentLoaded', (event) => {
    const clikeButtons = document.querySelectorAll('.clike-button');
    const cdislikeButtons = document.querySelectorAll('.cdislike-button');
  
    clikeButtons.forEach(button => {
        button.addEventListener('click', async (e) => {
            e.preventDefault(); // to stop page refresh
  
            const commentId = button.getAttribute('data-comment-id');
            const url = `/clike/${commentId}`;
  
            try {
                const response = await fetch(url, {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json'
                    },
                    body: JSON.stringify({ commentId })
                });
  
                if (response.ok) {
                    const result = await response.json();
  
                    // Update the like count in the DOM
                    const clikeCountSpan = document.getElementById(`clike-count-${commentId}`);
                    const cdislikeCountSpan = document.getElementById(`cdislike-count-${commentId}`);
                    clikeCountSpan.textContent = result.clikeCount;
                    cdislikeCountSpan.textContent = result.cdislikeCount;
  
                    // Update the heart colors
                    const heartLike = button.querySelector('.heartcLike');
                    const dislikeButton = document.querySelector(`.cdislike-button[data-comment-id="${commentId}"]`);
                    const heartDislike = dislikeButton.querySelector('.heartcDislike');
  
                    if (!heartLike.classList.contains('red')) {
                        heartLike.classList.add('red');
                        heartDislike.classList.remove('red');
                    } else {
                        heartLike.classList.remove('red');
                    }
                } else {
                    console.error("Form submission failed");
                }
            } catch (error) {
                console.error("Error submitting form:", error);
            }
        });
    });

    cdislikeButtons.forEach(button => {
        button.addEventListener('click', async (e) => {
            e.preventDefault();
  
            const commentId = button.getAttribute('data-comment-id');
            const url = `/cdislike/${commentId}`; // Changed postId to commentId
  
            try {
                const response = await fetch(url, {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json'
                    },
                    body: JSON.stringify({ commentId })
                });
  
                if (response.ok) {
                    const result = await response.json();
  
                    // Update the dislike count in the DOM
                    const dislikeCountSpan = document.getElementById(`cdislike-count-${commentId}`);
                    const likeCountSpan = document.getElementById(`clike-count-${commentId}`);
                    dislikeCountSpan.textContent = result.cdislikeCount;
                    likeCountSpan.textContent = result.clikeCount;
  
                    // Update the heart colors
                    const heartDislike = button.querySelector('.heartcDislike');
                    const likeButton = document.querySelector(`.clike-button[data-comment-id="${commentId}"]`);
                    const heartLike = likeButton.querySelector('.heartcLike');
  
                    if (!heartDislike.classList.contains('red')) {
                        heartDislike.classList.add('red');
                        heartLike.classList.remove('red');
                    } else {
                        heartDislike.classList.remove('red');
                    }
                } else {
                    console.error("Form submission failed");
                }
            } catch (error) {
                console.error("Error submitting form:", error);
            }
        });
    });
})
