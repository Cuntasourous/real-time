document.addEventListener('DOMContentLoaded', (event) => {
  const likeButtons = document.querySelectorAll('.like-button');
  const dislikeButtons = document.querySelectorAll('.dislike-button');

  likeButtons.forEach(button => {
      button.addEventListener('click', async (e) => {
          e.preventDefault(); // to stop page refresh

          const postId = button.getAttribute('data-post-id');
          const url = `/like/${postId}`;

          try {
              const response = await fetch(url, {
                  method: 'POST',
                  headers: {
                      'Content-Type': 'application/json'
                  },
                  body: JSON.stringify({ postId })
              });

              if (response.ok) {
                  const result = await response.json();

                  // Update the like count in the DOM
                  const likeCountSpan = document.getElementById(`like-count-${postId}`);
                  const dislikeCountSpan = document.getElementById(`dislike-count-${postId}`);
                  likeCountSpan.textContent = result.likeCount;
                  dislikeCountSpan.textContent = result.dislikeCount;

                  // Update the heart colors
                  const heartLike = button.querySelector('.heartLike');
                  const dislikeButton = document.querySelector(`.dislike-button[data-post-id="${postId}"]`);
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

          const postId = button.getAttribute('data-post-id');
          const url = `/dislike/${postId}`;

          try {
              const response = await fetch(url, {
                  method: 'POST',
                  headers: {
                      'Content-Type': 'application/json'
                  },
                  body: JSON.stringify({ postId })
              });

              if (response.ok) {
                  const result = await response.json();

                  // Update the dislike count in the DOM
                  const dislikeCountSpan = document.getElementById(`dislike-count-${postId}`);
                  const likeCountSpan = document.getElementById(`like-count-${postId}`);
                  dislikeCountSpan.textContent = result.dislikeCount;
                  likeCountSpan.textContent = result.likeCount;

                  // Update the heart colors
                  const heartDislike = button.querySelector('.heartDislike');
                  const likeButton = document.querySelector(`.like-button[data-post-id="${postId}"]`);
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
});
