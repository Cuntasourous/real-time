document.addEventListener('DOMContentLoaded', (event) => {
    attachCommentLikeListeners();
  });
  
  function attachCommentLikeListeners(){
    const likeButtons = document.querySelectorAll('.clike-button');
    const dislikeButtons = document.querySelectorAll('.cdislike-button');
  
    likeButtons.forEach(button => {
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
                    
                    const likeCountSpan = document.getElementById(`clike-count-${commentId}`);
                    const dislikeCountSpan = document.getElementById(`cdislike-count-${commentId}`);
                    likeCountSpan.textContent = result.likeCount;
                    dislikeCountSpan.textContent = result.dislikeCount;
  
                    const heartLike = button.querySelector('.heartLike');
                    const dislikeButton = document.querySelector(`.cdislike-button[data-comment-id="${commentId}"]`);
                    const heartDislike = dislikeButton.querySelector('.heartDislike');
  
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
  
    dislikeButtons.forEach(button => {
        button.addEventListener('click', async (e) => {
            e.preventDefault();
            const commentId = button.getAttribute('data-comment-id');
            const url = `/cdislike/${commentId}`;

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
                    dislikeCountSpan.textContent = result.dislikeCount;
                    likeCountSpan.textContent = result.likeCount;
  
                    // Update the heart colors
                    const heartDislike = button.querySelector('.heartDislike');
                    const likeButton = document.querySelector(`.clike-button[data-comment-id="${commentId}"]`);
                    const heartLike = likeButton.querySelector('.heartLike');
  
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
  }