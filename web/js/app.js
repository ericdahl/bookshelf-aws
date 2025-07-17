document.addEventListener('DOMContentLoaded', function() {
    // Check authentication before loading app
    if (!checkAuthentication()) {
        window.location.href = 'signin.html';
        return;
    }

    // API endpoints
    const API_BASE_URL = 'https://wl32jdoac6.execute-api.us-east-1.amazonaws.com/prod';
    const API = {
        BOOKS: `${API_BASE_URL}/books`,
        SEARCH: `${API_BASE_URL}/search`, // Google Books search endpoint
        BOOK_STATUS: (id) => `${API_BASE_URL}/books/${id}`,
        BOOK_DETAILS: (id) => `${API_BASE_URL}/books/${id}`,
        DELETE_BOOK: (id) => `${API_BASE_URL}/books/${id}`
    };

    // DOM Elements
    const searchInput = document.getElementById('search-input');
    const searchButton = document.getElementById('search-button');
    const searchCollectionButton = document.getElementById('search-collection-button');
    const searchResults = document.getElementById('search-results');
    const closeSearch = document.getElementById('close-search');
    const resultsContainer = document.querySelector('.results-container');
    const shelves = document.querySelectorAll('.books-container');
    const bookDetails = document.getElementById('book-details');
    const closeDetails = document.getElementById('close-details');
    const saveDetails = document.getElementById('save-details');
    const deleteBookButton = document.getElementById('delete-book');
    const loadingOverlay = document.getElementById('loading-overlay');
    const ratingStars = document.querySelectorAll('.stars i');
    const fullViewButton = document.getElementById('full-view');
    const compactViewButton = document.getElementById('compact-view');
    const shelvesContainer = document.querySelector('.shelves-container');

    // Current book being viewed/edited
    let currentBook = null;
    let currentRating = null;

    // Initialize the application
    initApp();

    // Initialize the application
    function initApp() {
        // Load all books from the server
        loadBooks();

        // Set up event listeners
        setupEventListeners();

        // Initialize drag and drop
        initDragAndDrop();
        
        // Initialize shelf sorting
        initShelfSorting();
    }
    
    // Initialize shelf sorting
    function initShelfSorting() {
        // Default sort by title for all shelves when app starts
        document.querySelectorAll('.books-container').forEach(container => {
            const status = container.dataset.status;
            if (status) {
                sortShelfBooks(status, 'title', 'asc');
            }
        });
    }
    
    // Handle column header clicks for sorting
    function handleHeaderClick(event) {
        const header = event.currentTarget;
        const sortBy = header.dataset.sort;
        const status = header.dataset.status;
        
        // Determine sort direction
        let sortDirection = 'asc';
        if (header.classList.contains('sorted-asc')) {
            sortDirection = 'desc';
        }
        
        // Sort the books
        sortShelfBooks(status, sortBy, sortDirection);
    }

    // Load all books from the server
    function loadBooks() {
        showLoading();
        fetch(API.BOOKS, {
            headers: getAuthHeaders()
        })
            .then(response => response.json())
            .then(books => {
                // Clear existing books from shelves
                document.querySelectorAll('.books-container').forEach(shelf => {
                    shelf.innerHTML = '';
                });
                
                // Group books by status
                const booksByStatus = {
                    'Want to Read': [],
                    'Currently Reading': [],
                    'Read': []
                };
                
                // Map backend status to frontend status
                const statusMapping = {
                    'WANT_TO_READ': 'Want to Read',
                    'READING': 'Currently Reading', 
                    'READ': 'Read'
                };
                
                books.forEach(book => {
                    const frontendStatus = statusMapping[book.status] || book.status;
                    if (booksByStatus[frontendStatus]) {
                        // Update the book object with the frontend status
                        book.status = frontendStatus;
                        booksByStatus[frontendStatus].push(book);
                    }
                });
                
                // Sort each shelf's books by title (default) and add to shelf
                Object.keys(booksByStatus).forEach(status => {
                    const sortSelect = document.querySelector(`.sort-select[data-status="${status}"]`);
                    const sortBy = sortSelect ? sortSelect.value : 'title';
                    
                    const sortedBooks = sortBooks(booksByStatus[status], sortBy);
                    sortedBooks.forEach(book => {
                        addBookToShelf(book);
                    });
                });
                
                hideLoading();
            })
            .catch(error => {
                console.error('Error loading books:', error);
                hideLoading();
                alert('Failed to load books. Please try again.');
            });
    }
    
    // Sort an array of books by the given criteria
    function sortBooks(books, sortBy) {
        return [...books].sort((a, b) => {
            switch (sortBy) {
                case 'title':
                    return a.title.localeCompare(b.title);
                case 'author':
                    return a.author.localeCompare(b.author);
                case 'rating':
                    // Handle null ratings (null ratings go at the end)
                    if (a.rating === null && b.rating === null) return 0;
                    if (a.rating === null) return 1;
                    if (b.rating === null) return -1;
                    // Sort by rating in descending order (higher ratings first)
                    return b.rating - a.rating;
                default:
                    return 0;
            }
        });
    }
    
    // Sort books in a specific shelf with ascending/descending order support
    function sortShelfBooks(status, sortBy, sortDirection = 'asc') {
        // Get all books from this shelf
        const shelf = document.querySelector(`.books-container[data-status="${status}"]`);
        if (!shelf) return;
        
        // Get all book cards in this shelf
        const bookCards = Array.from(shelf.querySelectorAll('.book-card'));
        
        // Map to objects with data for sorting
        const booksData = bookCards.map(card => {
            const id = card.dataset.id;
            const title = card.querySelector('.book-title').textContent;
            const author = card.querySelector('.book-author').textContent;
            
            // Get series data if it exists
            let series = "";
            const seriesElem = card.querySelector('.book-series');
            if (seriesElem && seriesElem.textContent !== '-') {
                series = seriesElem.textContent;
            }
            
            // Parse rating if it exists
            let rating = null;
            const ratingElem = card.querySelector('.book-rating');
            if (ratingElem) {
                const ratingMatch = ratingElem.textContent.match(/(\d+)/);
                if (ratingMatch) {
                    rating = parseInt(ratingMatch[1], 10);
                }
            }
            
            return { element: card, id, title, author, series, rating };
        });
        
        // Sort books
        const sortedBooks = [...booksData].sort((a, b) => {
            let result;
            
            switch (sortBy) {
                case 'title':
                    result = a.title.localeCompare(b.title);
                    break;
                case 'author':
                    result = a.author.localeCompare(b.author);
                    break;
                case 'series':
                    // Handle empty series (empty series go at the end)
                    if (a.series === "" && b.series === "") return 0;
                    if (a.series === "") return 1;
                    if (b.series === "") return -1;
                    result = a.series.localeCompare(b.series);
                    break;
                case 'rating':
                    // Handle null ratings (null ratings go at the end)
                    if (a.rating === null && b.rating === null) return 0;
                    if (a.rating === null) return 1;
                    if (b.rating === null) return -1;
                    // Sort by rating
                    result = a.rating - b.rating;
                    break;
                default:
                    result = 0;
            }
            
            // Apply sort direction
            return sortDirection === 'asc' ? result : -result;
        });
        
        // Remove all books from shelf
        bookCards.forEach(card => card.remove());
        
        // Add back in sorted order
        sortedBooks.forEach(book => {
            shelf.appendChild(book.element);
        });
        
        // Update header sort indicators
        // In compact mode, headers are within books-container-header as the first child of the shelf
        const headerContainer = shelf.querySelector('.books-container-header');
        if (headerContainer) {
            const headers = headerContainer.querySelectorAll('.header-cell');
            headers.forEach(header => {
                // Remove sort indicators from all headers
                header.classList.remove('sorted-asc', 'sorted-desc');
                
                // Add appropriate sort indicator to the active sort header
                if (header.dataset.sort === sortBy) {
                    header.classList.add(sortDirection === 'asc' ? 'sorted-asc' : 'sorted-desc');
                }
            });
        }
    }

    // Set up event listeners
    function setupEventListeners() {
        // Search
        searchButton.addEventListener('click', searchGoogleBooks);
        searchCollectionButton.addEventListener('click', searchCollectionBooks);
        searchInput.addEventListener('keypress', e => {
            if (e.key === 'Enter') {
                searchGoogleBooks();
            }
        });
        
        // Close search results
        closeSearch.addEventListener('click', () => {
            searchResults.classList.add('hidden');
            searchInput.value = ''; // Clear search input
        });

        // Book details
        closeDetails.addEventListener('click', () => {
            bookDetails.classList.add('hidden');
        });
        
        // Rating stars
        ratingStars.forEach(star => {
            star.addEventListener('click', function() {
                const rating = parseInt(this.dataset.rating);
                updateRatingUI(rating);
                currentRating = rating;
            });
            
            // Add hover effect
            star.addEventListener('mouseenter', function() {
                const rating = parseInt(this.dataset.rating);
                previewRating(rating);
            });
        });
        
        // Reset rating preview on mouseleave
        document.querySelector('.stars').addEventListener('mouseleave', function() {
            updateRatingUI(currentRating || 0);
        });
        
        // Save book details
        saveDetails.addEventListener('click', saveBookDetails);
        
        // Delete book
        deleteBookButton.addEventListener('click', deleteBook);
        
        // View toggle buttons
        fullViewButton.addEventListener('click', () => {
            setViewMode('full');
        });
        
        compactViewButton.addEventListener('click', () => {
            setViewMode('compact');
        });
        
        // Load saved view preference
        loadViewPreference();
    }
    
    // Set the view mode (full or compact)
    function setViewMode(mode) {
        if (mode === 'compact') {
            shelvesContainer.classList.add('compact-mode');
            fullViewButton.classList.remove('active');
            compactViewButton.classList.add('active');
            // Save preference
            localStorage.setItem('bookshelfViewMode', 'compact');
            
            // Add table headers to each shelf
            document.querySelectorAll('.books-container').forEach(container => {
                // Remove existing headers if any
                const existingHeader = container.querySelector('.books-container-header');
                if (existingHeader) {
                    existingHeader.remove();
                }
                
                // Get the shelf status
                const status = container.dataset.status;
                
                // Create new header
                const header = document.createElement('div');
                header.className = 'books-container-header';
                
                const headerRow = document.createElement('div');
                headerRow.className = 'header-row';
                
                // Add column headers
                const titleHeader = document.createElement('div');
                titleHeader.className = 'header-cell cell-title sorted-asc'; // Default sort column
                titleHeader.textContent = 'Title';
                titleHeader.dataset.sort = 'title';
                titleHeader.dataset.status = status;
                titleHeader.addEventListener('click', handleHeaderClick);
                
                const authorHeader = document.createElement('div');
                authorHeader.className = 'header-cell cell-author';
                authorHeader.textContent = 'Author';
                authorHeader.dataset.sort = 'author';
                authorHeader.dataset.status = status;
                authorHeader.addEventListener('click', handleHeaderClick);
                
                const seriesHeader = document.createElement('div');
                seriesHeader.className = 'header-cell cell-series';
                seriesHeader.textContent = 'Series';
                seriesHeader.dataset.sort = 'series';
                seriesHeader.dataset.status = status;
                seriesHeader.addEventListener('click', handleHeaderClick);
                
                const ratingHeader = document.createElement('div');
                ratingHeader.className = 'header-cell cell-rating';
                ratingHeader.textContent = 'Rating';
                ratingHeader.dataset.sort = 'rating';
                ratingHeader.dataset.status = status;
                ratingHeader.addEventListener('click', handleHeaderClick);
                
                // Assemble header
                headerRow.appendChild(titleHeader);
                headerRow.appendChild(authorHeader);
                headerRow.appendChild(seriesHeader);
                headerRow.appendChild(ratingHeader);
                header.appendChild(headerRow);
                
                // Insert header at the beginning of the container
                container.insertBefore(header, container.firstChild);
                
                // Convert existing book cards to tabular format
                container.querySelectorAll('.book-card').forEach(convertBookCardToTableRow);
            });
        } else {
            shelvesContainer.classList.remove('compact-mode');
            fullViewButton.classList.add('active');
            compactViewButton.classList.remove('active');
            // Save preference
            localStorage.setItem('bookshelfViewMode', 'full');
            
            // Remove table headers
            document.querySelectorAll('.books-container-header').forEach(header => {
                header.remove();
            });
            
            // Restore original book card structure if needed
            document.querySelectorAll('.book-card').forEach(card => {
                // Make sure book-info is displayed
                const infoDiv = card.querySelector('.book-info');
                if (infoDiv) {
                    infoDiv.style.removeProperty('display');
                }
                
                // Remove any table cells if they exist
                const cells = card.querySelectorAll('.cell-title, .cell-author, .cell-series, .cell-rating');
                cells.forEach(cell => cell.remove());
            });
        }
    }
    
    // Convert a book card to a table row format
    function convertBookCardToTableRow(card) {
        // If cells already exist, just return
        if (card.querySelector('.cell-title')) {
            return;
        }
        
        // Get book data from existing elements
        const title = card.querySelector('.book-title').textContent;
        const author = card.querySelector('.book-author').textContent;
        
        // Check if book is an audiobook
        const typeElement = card.querySelector('.book-type');
        const isAudiobook = typeElement && typeElement.textContent.includes('Audiobook');
        
        // Create table cells
        const titleCell = document.createElement('div');
        titleCell.className = 'cell-title';
        const typeIcon = isAudiobook ? '<i class="fas fa-headphones"></i> ' : '';
        titleCell.innerHTML = `<div class="book-title">${typeIcon}${title}</div>`;
        
        const authorCell = document.createElement('div');
        authorCell.className = 'cell-author';
        authorCell.innerHTML = `<div class="book-author">${author}</div>`;
        
        const seriesCell = document.createElement('div');
        seriesCell.className = 'cell-series';
        const seriesElement = card.querySelector('.book-series');
        if (seriesElement) {
            seriesCell.innerHTML = `<div class="book-series">${seriesElement.textContent}</div>`;
        } else {
            seriesCell.innerHTML = `<div class="book-series">-</div>`;
        }
        
        const ratingCell = document.createElement('div');
        ratingCell.className = 'cell-rating';
        const ratingElement = card.querySelector('.book-rating');
        if (ratingElement) {
            ratingCell.innerHTML = `<div class="book-rating">${ratingElement.textContent.replace('Rating: ', '')}</div>`;
        } else {
            ratingCell.innerHTML = `<div class="book-rating">-</div>`;
        }
        
        // Add cells to the card
        card.appendChild(titleCell);
        card.appendChild(authorCell);
        card.appendChild(seriesCell);
        card.appendChild(ratingCell);
    }
    
    // Load saved view preference
    function loadViewPreference() {
        const savedMode = localStorage.getItem('bookshelfViewMode');
        
        // We need to ensure loadBooks completes before applying view mode
        // This ensures headers are properly added to populated shelves
        const observer = new MutationObserver((mutations) => {
            mutations.forEach((mutation) => {
                if (mutation.type === 'childList' && mutation.addedNodes.length > 0) {
                    // Books have been added to shelves
                    if (savedMode === 'compact') {
                        setViewMode('compact');
                    } else {
                        setViewMode('full'); // Default to full view
                    }
                    observer.disconnect();
                }
            });
        });
        
        // Start observing all book containers
        document.querySelectorAll('.books-container').forEach(container => {
            observer.observe(container, { childList: true });
        });
        
        // Set a timeout in case no books are loaded
        setTimeout(() => {
            if (savedMode === 'compact') {
                setViewMode('compact');
            } else {
                setViewMode('full');
            }
            observer.disconnect();
        }, 1000);
    }

    // Initialize drag and drop
    function initDragAndDrop() {
        shelves.forEach(shelf => {
            new Sortable(shelf, {
                group: 'books',
                animation: 150,
                ghostClass: 'sortable-ghost',
                dragClass: 'sortable-drag',
                onEnd: function(evt) {
                    const bookId = evt.item.dataset.id;
                    const newStatus = evt.to.dataset.status;
                    
                    // Update the book status on the server
                    updateBookStatus(bookId, newStatus);
                }
            });
        });
    }

    // Search for books using Google Books API
    function searchGoogleBooks() {
        const query = searchInput.value.trim();
        
        if (query === '') {
            return;
        }
        
        showLoading();
        
        fetch(`${API.SEARCH}?q=${encodeURIComponent(query)}`, {
            headers: getAuthHeaders()
        })
            .then(response => {
                if (!response.ok) {
                    throw new Error('Failed to search books');
                }
                return response.json();
            })
            .then(books => {
                // Get existing books to check duplicates
                const existingBooks = getAllExistingBooks();
                
                // Check which books already exist in our collection
                const searchResultsWithStatus = books.map(book => {
                    const existingBook = existingBooks.find(existing => 
                        existing.title.toLowerCase() === book.title.toLowerCase() &&
                        existing.author.toLowerCase() === book.author.toLowerCase()
                    );
                    
                    if (existingBook) {
                        book.existing_shelf = existingBook.status;
                    }
                    
                    return book;
                });
                
                displaySearchResults(searchResultsWithStatus, `Google Books results for "${query}"`);
                hideLoading();
            })
            .catch(error => {
                console.error('Error searching books:', error);
                hideLoading();
                alert('Failed to search for books. Please try again.');
            });
    }

    // Search for books in existing collection
    function searchCollectionBooks() {
        const query = searchInput.value.trim();
        
        if (query === '') {
            return;
        }
        
        showLoading();
        
        // Get all books from all shelves
        const allBooks = [];
        document.querySelectorAll('.book-card').forEach(card => {
            const title = card.querySelector('.book-title').textContent;
            const author = card.querySelector('.book-author').textContent;
            const id = card.dataset.id;
            
            allBooks.push({
                id: id,
                title: title,
                author: author,
                element: card
            });
        });
        
        // Filter books based on search query
        const filteredBooks = allBooks.filter(book => 
            book.title.toLowerCase().includes(query.toLowerCase()) ||
            book.author.toLowerCase().includes(query.toLowerCase())
        );
        
        displayCollectionSearchResults(filteredBooks, query);
        hideLoading();
    }

    // Get all existing books from the shelves
    function getAllExistingBooks() {
        const allBooks = [];
        document.querySelectorAll('.book-card').forEach(card => {
            const title = card.querySelector('.book-title').textContent;
            const author = card.querySelector('.book-author').textContent;
            const id = card.dataset.id;
            
            // Get the shelf/status
            const shelf = card.closest('.books-container');
            const status = shelf ? shelf.dataset.status : '';
            
            allBooks.push({
                id: id,
                title: title,
                author: author,
                status: status
            });
        });
        return allBooks;
    }

    // Display search results from Google Books API
    function displaySearchResults(books, headerText) {
        resultsContainer.innerHTML = '';
        
        if (books.length === 0) {
            resultsContainer.innerHTML = '<p>No books found. Try a different search term.</p>';
        } else {
            // Show number of results
            resultsContainer.innerHTML = `<p class="search-count">${books.length} ${headerText}</p>`;
            
            // Create a container for the book cards
            const booksGrid = document.createElement('div');
            booksGrid.className = 'search-results-grid';
            
            // Add each book to the grid
            books.forEach(book => {
                const bookCard = createSearchResultCard(book);
                booksGrid.appendChild(bookCard);
            });
            
            resultsContainer.appendChild(booksGrid);
        }
        
        // Automatically scroll to the search results
        searchResults.classList.remove('hidden');
        searchResults.scrollIntoView({ behavior: 'smooth' });
    }

    // Display search results from collection
    function displayCollectionSearchResults(filteredBooks, query) {
        resultsContainer.innerHTML = '';
        
        if (filteredBooks.length === 0) {
            resultsContainer.innerHTML = '<p>No books found in your collection. Try a different search term.</p>';
        } else {
            // Show number of results
            resultsContainer.innerHTML = `<p class="search-count">${filteredBooks.length} books found in your collection for "${query}"</p>`;
            
            // Create a container for the book cards
            const booksGrid = document.createElement('div');
            booksGrid.className = 'search-results-grid';
            
            // Add each book to the grid (clone the existing cards)
            filteredBooks.forEach(book => {
                const clonedCard = book.element.cloneNode(true);
                clonedCard.classList.add('search-result');
                booksGrid.appendChild(clonedCard);
            });
            
            resultsContainer.appendChild(booksGrid);
        }
        
        // Automatically scroll to the search results
        searchResults.classList.remove('hidden');
        searchResults.scrollIntoView({ behavior: 'smooth' });
    }

    // Add a book to the appropriate shelf
    function addBookToShelf(book) {
        const shelf = document.querySelector(`.books-container[data-status="${book.status}"]`);
        if (shelf) {
            const bookCard = createBookCard(book);
            shelf.appendChild(bookCard);
            
            // If in compact mode, convert the card to table row format
            if (shelvesContainer.classList.contains('compact-mode')) {
                convertBookCardToTableRow(bookCard);
            }
        }
    }

    // Create a book card element
    function createBookCard(book) {
        const card = document.createElement('div');
        card.className = 'book-card';
        card.dataset.id = book.id;
        
        const coverUrl = book.thumbnail || book.cover_url || 'https://via.placeholder.com/150x200?text=No+Cover';
        const ratingHtml = book.rating ? `<p class="book-rating">Rating: ${book.rating}/10</p>` : '';
        
        // Prepare series info display if available
        let seriesHtml = '';
        if (book.series && book.series_index) {
            seriesHtml = `<p class="book-series">${book.series} Book ${book.series_index}</p>`;
        } else if (book.series) {
            seriesHtml = `<p class="book-series">${book.series}</p>`;
        }
        
        // Show book type if it's an audiobook (default type "book" isn't shown to keep UI clean)
        const typeHtml = book.type === 'audiobook' ? `<p class="book-type"><i class="fas fa-headphones"></i> Audiobook</p>` : '';
        
        card.innerHTML = `
            <div class="book-cover">
                <img src="${coverUrl}" alt="${book.title} cover">
            </div>
            <div class="book-info">
                <h3 class="book-title">${book.title}</h3>
                <p class="book-author">${book.author}</p>
                ${seriesHtml}
                ${typeHtml}
                ${ratingHtml}
            </div>
        `;
        
        // Add click event to open book details
        card.addEventListener('click', () => {
            showBookDetails(book);
        });
        
        return card;
    }

    // Create a search result card
    function createSearchResultCard(book) {
        const card = document.createElement('div');
        card.className = 'book-card search-result';
        
        const coverUrl = book.thumbnail || 'https://via.placeholder.com/150x200?text=No+Cover';
        
        // Set button text based on whether book is already in a shelf
        let buttonText = "Add to Want to Read";
        let buttonClass = "add-book";
        if (book.existing_shelf) {
            buttonText = `In ${book.existing_shelf}`;
            buttonClass = "add-book book-exists";
        }

        card.innerHTML = `
            <div class="book-cover">
                <img src="${coverUrl}" alt="${book.title} cover">
            </div>
            <div class="book-info">
                <h3 class="book-title">${book.title}</h3>
                <p class="book-author">${book.author}</p>
                <button class="${buttonClass}">${buttonText}</button>
            </div>
        `;
        
        // Add event listener to the add button (only if it's not already in a shelf)
        const addButton = card.querySelector('.add-book');
        if (!book.existing_shelf) {
            addButton.addEventListener('click', () => {
                addGoogleBook(book, addButton);
            });
        }
        
        return card;
    }

    // Add a new book from Google Books search to the shelf
    function addGoogleBook(book, buttonElement) {
        // If book already exists in a shelf, just show a notification
        if (book.existing_shelf) {
            alert(`This book is already in your "${book.existing_shelf}" shelf`);
            return;
        }
        
        showLoading();
        
        const newBook = {
            title: book.title,
            author: book.author,
            status: 'WANT_TO_READ', // Use backend status format
            thumbnail: book.thumbnail || ''
        };
        
        fetch(API.BOOKS, {
            method: 'POST',
            headers: getAuthHeaders(),
            body: JSON.stringify(newBook)
        })
        .then(response => {
            if (!response.ok) {
                throw new Error('Failed to add book');
            }
            return response.json();
        })
        .then(addedBook => {
            // Map the status to frontend format for display
            const statusMapping = {
                'WANT_TO_READ': 'Want to Read',
                'READING': 'Currently Reading',
                'read': 'Read'
            };
            addedBook.status = statusMapping[addedBook.status] || addedBook.status;
            
            // Add the book to the shelf
            addBookToShelf(addedBook);
            
            // Update button to show it was added
            if (buttonElement) {
                buttonElement.textContent = "Added ✓";
                buttonElement.disabled = true;
                buttonElement.classList.add("book-added");
            }
            
            // Keep search results open for adding more books
            hideLoading();
        })
        .catch(error => {
            console.error('Error adding book:', error);
            hideLoading();
            alert('Failed to add book. Please try again.');
        });
    }

    // Update a book's status (currently not implemented in backend)
    function updateBookStatus(bookId, newStatus) {
        // Map frontend status to backend status
        const statusMapping = {
            'Want to Read': 'WANT_TO_READ',
            'Currently Reading': 'READING',
            'Read': 'READ'
        };
        
        const backendStatus = statusMapping[newStatus] || newStatus;
        
        showLoading();
        
        fetch(`${API_BASE_URL}/books/${bookId}`, {
            method: 'PUT',
            headers: getAuthHeaders(),
            body: JSON.stringify({
                status: backendStatus
            })
        })
        .then(response => {
            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }
            return response.json();
        })
        .then(updatedBook => {
            // Update the book in the UI
            updatedBook.status = newStatus; // Use frontend status for UI
            updateBookCardInShelf(updatedBook);
            hideLoading();
        })
        .catch(error => {
            console.error('Error updating book status:', error);
            hideLoading();
            alert('Failed to update book status. Please try again.');
            // Reload books to reset the UI to the server state
            loadBooks();
        });
    }

    // Show book details
    function showBookDetails(book) {
        currentBook = book;
        
        // Update the UI with book details
        document.getElementById('detail-title').textContent = book.title;
        document.getElementById('detail-author').textContent = book.author;
        document.getElementById('detail-cover').src = book.thumbnail || book.cover_url || 'https://via.placeholder.com/150x200?text=No+Cover';
        
        // Update OpenLibrary link
        const openLibraryLink = document.getElementById('detail-openlibrary-link').querySelector('a');
        if (book.open_library_id) {
            // Check if the ID is in the format OL12345M or if it's a full path like /works/OL12345M
            let olid = book.open_library_id;
            if (olid.startsWith('/')) {
                // Extract just the ID part
                const parts = olid.split('/');
                olid = parts[parts.length - 1];
            }
            
            // Set the URL based on the format of the ID
            let url;
            if (olid.startsWith('OL') && olid.endsWith('M')) {
                // It's an edition ID (starts with OL and ends with M)
                url = `https://openlibrary.org/books/${olid}`;
            } else if (olid.startsWith('OL') && olid.endsWith('W')) {
                // It's a works ID (starts with OL and ends with W)
                url = `https://openlibrary.org/works/${olid}`;
            } else {
                // Default to works path 
                url = `https://openlibrary.org/works/${olid}`;
            }
            
            openLibraryLink.href = url;
            openLibraryLink.parentElement.style.display = 'block';
        } else {
            openLibraryLink.parentElement.style.display = 'none';
        }
        
        // Update rating UI
        updateRatingUI(book.rating || 0);
        currentRating = book.rating || null;
        
        // Update type radio buttons
        document.getElementById('type-book').checked = book.type === 'book' || !book.type;
        document.getElementById('type-audiobook').checked = book.type === 'audiobook';
        
        // Update comments
        document.getElementById('book-comments').value = book.comments || '';
        
        // Update series information
        document.getElementById('book-series').value = book.series || '';
        document.getElementById('book-series-index').value = book.series_index || '';
        
        // Show the details popup
        bookDetails.classList.remove('hidden');
    }

    // Preview rating on hover
    function previewRating(rating) {
        ratingStars.forEach((star, index) => {
            if (index < rating) {
                star.className = 'fas fa-star';
            } else {
                star.className = 'far fa-star';
            }
        });
        
        // Update the rating value text
        const ratingText = rating > 0 ? rating.toString() : "None";
        document.getElementById('rating-value').textContent = ratingText;
    }
    
    // Update rating UI
    function updateRatingUI(rating) {
        ratingStars.forEach((star, index) => {
            if (index < rating) {
                star.className = 'fas fa-star';
            } else {
                star.className = 'far fa-star';
            }
        });
        
        // Update the rating value text
        const ratingText = rating > 0 ? rating.toString() : "None";
        document.getElementById('rating-value').textContent = ratingText;
    }

    // Save book details
    function saveBookDetails() {
        if (!currentBook) return;
        
        // Get the updated values from the form
        const updatedBook = {
            series: document.getElementById('book-series').value || undefined,
            rating: currentRating || undefined,
            review: document.getElementById('book-comments').value || undefined
        };
        
        // Only include fields that have values
        Object.keys(updatedBook).forEach(key => {
            if (updatedBook[key] === undefined || updatedBook[key] === '') {
                delete updatedBook[key];
            }
        });
        
        // If no changes, just close the dialog
        if (Object.keys(updatedBook).length === 0) {
            bookDetails.classList.add('hidden');
            return;
        }
        
        showLoading();
        
        fetch(`${API_BASE_URL}/books/${currentBook.id}`, {
            method: 'PUT',
            headers: getAuthHeaders(),
            body: JSON.stringify(updatedBook)
        })
        .then(response => {
            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }
            return response.json();
        })
        .then(updatedBookData => {
            // Map backend status to frontend status for UI
            const statusMapping = {
                'WANT_TO_READ': 'Want to Read',
                'READING': 'Currently Reading',
                'READ': 'Read'
            };
            updatedBookData.status = statusMapping[updatedBookData.status] || updatedBookData.status;
            
            // Update the book in the UI
            updateBookCardInShelf(updatedBookData);
            hideLoading();
            
            // Close the details popup
            bookDetails.classList.add('hidden');
        })
        .catch(error => {
            console.error('Error updating book details:', error);
            hideLoading();
            alert('Failed to update book details. Please try again.');
        });
    }

    // Delete a book
    function deleteBook() {
        if (!currentBook || !confirm(`Are you sure you want to delete "${currentBook.title}"? This action cannot be undone.`)) return;
        
        showLoading();
        
        fetch(`${API_BASE_URL}/books/${currentBook.id}`, {
            method: 'DELETE',
            headers: getAuthHeaders()
        })
        .then(response => {
            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }
            // No need to parse JSON for 204 No Content response
            return response.status;
        })
        .then(() => {
            // Remove the book from the UI
            const bookCard = document.querySelector(`.book-card[data-id="${currentBook.id}"]`);
            if (bookCard) {
                bookCard.remove();
            }
            
            hideLoading();
            
            // Close the details popup
            bookDetails.classList.add('hidden');
            
            // Clear current book reference
            currentBook = null;
        })
        .catch(error => {
            console.error('Error deleting book:', error);
            hideLoading();
            alert('Failed to delete book. Please try again.');
        });
    }

    // Show loading overlay
    function showLoading() {
        loadingOverlay.classList.remove('hidden');
    }

    // Hide loading overlay
    function hideLoading() {
        loadingOverlay.classList.add('hidden');
    }

    // Update book card in shelf
    function updateBookCardInShelf(book) {
        const bookCard = document.querySelector(`.book-card[data-id="${book.id}"]`);
        if (bookCard) {
            // Check if we're in compact view
            const isCompactMode = shelvesContainer.classList.contains('compact-mode');
            
            // Update the standard card view elements
            const bookInfo = bookCard.querySelector('.book-info');
            if (bookInfo) {
                // Update rating if needed
                let ratingElement = bookCard.querySelector('.book-info .book-rating');
                if (book.rating) {
                    if (ratingElement) {
                        ratingElement.textContent = `Rating: ${book.rating}/10`;
                    } else {
                        const ratingP = document.createElement('p');
                        ratingP.className = 'book-rating';
                        ratingP.textContent = `Rating: ${book.rating}/10`;
                        bookInfo.appendChild(ratingP);
                    }
                } else if (ratingElement) {
                    ratingElement.remove();
                }
                
                // Update type info if needed
                let typeElement = bookCard.querySelector('.book-info .book-type');
                if (book.type === 'audiobook') {
                    if (typeElement) {
                        typeElement.innerHTML = `<i class="fas fa-headphones"></i> Audiobook`;
                    } else {
                        const authorElement = bookInfo.querySelector('.book-author');
                        
                        const typeP = document.createElement('p');
                        typeP.className = 'book-type';
                        typeP.innerHTML = `<i class="fas fa-headphones"></i> Audiobook`;
                        
                        // Insert after author element or series element if it exists
                        const seriesElement = bookInfo.querySelector('.book-series');
                        const insertAfter = seriesElement || authorElement;
                        
                        if (insertAfter.nextSibling) {
                            bookInfo.insertBefore(typeP, insertAfter.nextSibling);
                        } else {
                            bookInfo.appendChild(typeP);
                        }
                    }
                } else if (typeElement) {
                    typeElement.remove();
                }
                
                // Update series info if needed
                let seriesElement = bookCard.querySelector('.book-info .book-series');
                if (book.series) {
                    const seriesText = book.series_index 
                        ? `${book.series} Book ${book.series_index}` 
                        : book.series;
                        
                    if (seriesElement) {
                        seriesElement.textContent = seriesText;
                    } else {
                        const authorElement = bookInfo.querySelector('.book-author');
                        
                        const seriesP = document.createElement('p');
                        seriesP.className = 'book-series';
                        seriesP.textContent = seriesText;
                        
                        // Insert after author element
                        if (authorElement.nextSibling) {
                            bookInfo.insertBefore(seriesP, authorElement.nextSibling);
                        } else {
                            bookInfo.appendChild(seriesP);
                        }
                    }
                } else if (seriesElement) {
                    seriesElement.remove();
                }
            }
            
            // Update compact view cells if in compact mode
            if (isCompactMode) {
                // Update rating cell
                const ratingCell = bookCard.querySelector('.cell-rating');
                if (ratingCell) {
                    if (book.rating) {
                        ratingCell.innerHTML = `<div class="book-rating">${book.rating}/10</div>`;
                    } else {
                        ratingCell.innerHTML = `<div class="book-rating">-</div>`;
                    }
                }
                
                // Update series cell
                const seriesCell = bookCard.querySelector('.cell-series');
                if (seriesCell) {
                    if (book.series) {
                        const seriesText = book.series_index 
                            ? `${book.series} Book ${book.series_index}` 
                            : book.series;
                        seriesCell.innerHTML = `<div class="book-series">${seriesText}</div>`;
                    } else {
                        seriesCell.innerHTML = `<div class="book-series">-</div>`;
                    }
                }
                
                // Update title cell with type icon for audiobooks
                const titleCell = bookCard.querySelector('.cell-title');
                if (titleCell) {
                    const titleText = book.title;
                    const typeIcon = book.type === 'audiobook' ? '<i class="fas fa-headphones"></i> ' : '';
                    titleCell.innerHTML = `<div class="book-title">${typeIcon}${titleText}</div>`;
                }
            }
        }
    }

    // Authentication functions
    function checkAuthentication() {
        const accessToken = localStorage.getItem('accessToken');
        const refreshToken = localStorage.getItem('refreshToken');
        
        if (!accessToken || !refreshToken) {
            return false;
        }
        
        // Check if access token is expired
        try {
            const tokenPayload = JSON.parse(atob(accessToken.split('.')[1]));
            const currentTime = Math.floor(Date.now() / 1000);
            
            if (tokenPayload.exp < currentTime) {
                // Token is expired, try to refresh
                return refreshAccessToken();
            }
            
            return true;
        } catch (error) {
            console.error('Error validating token:', error);
            clearAuthTokens();
            return false;
        }
    }

    async function refreshAccessToken() {
        const refreshToken = localStorage.getItem('refreshToken');
        
        if (!refreshToken) {
            return false;
        }
        
        try {
            // Initialize Cognito SDK if not already done
            const poolData = {
                UserPoolId: 'us-east-1_yqnoWMmU6',
                ClientId: '4645dqoa4ng95qqn9kkeearsil'
            };
            
            const userPool = new AmazonCognitoIdentity.CognitoUserPool(poolData);
            const cognitoUser = userPool.getCurrentUser();
            
            if (!cognitoUser) {
                clearAuthTokens();
                return false;
            }
            
            return new Promise((resolve) => {
                cognitoUser.getSession((err, session) => {
                    if (err) {
                        console.error('Error refreshing token:', err);
                        clearAuthTokens();
                        resolve(false);
                        return;
                    }
                    
                    if (session.isValid()) {
                        // Update stored tokens
                        localStorage.setItem('accessToken', session.getAccessToken().getJwtToken());
                        localStorage.setItem('idToken', session.getIdToken().getJwtToken());
                        localStorage.setItem('refreshToken', session.getRefreshToken().getToken());
                        resolve(true);
                    } else {
                        clearAuthTokens();
                        resolve(false);
                    }
                });
            });
        } catch (error) {
            console.error('Error refreshing token:', error);
            clearAuthTokens();
            return false;
        }
    }

    function clearAuthTokens() {
        localStorage.removeItem('accessToken');
        localStorage.removeItem('idToken');
        localStorage.removeItem('refreshToken');
        localStorage.removeItem('userEmail');
    }

    function signOut() {
        // Clear authentication state
        clearAuthTokens();
        
        // Sign out from Cognito
        try {
            const poolData = {
                UserPoolId: 'us-east-1_yqnoWMmU6',
                ClientId: '4645dqoa4ng95qqn9kkeearsil'
            };
            
            const userPool = new AmazonCognitoIdentity.CognitoUserPool(poolData);
            const cognitoUser = userPool.getCurrentUser();
            
            if (cognitoUser) {
                cognitoUser.signOut();
            }
        } catch (error) {
            console.error('Error signing out:', error);
        }
        
        // Redirect to signin page
        window.location.href = 'signin.html';
    }

    // Function to get authorization headers for API requests
    function getAuthHeaders() {
        const accessToken = localStorage.getItem('accessToken');
        return accessToken ? {
            'Authorization': `Bearer ${accessToken}`,
            'Content-Type': 'application/json'
        } : {
            'Content-Type': 'application/json'
        };
    }

    // Make signOut function available globally
    window.signOut = signOut;
});