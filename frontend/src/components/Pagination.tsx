import React from 'react';

interface PaginationProps {
  currentPage: number;
  totalTasks: number;
  limit: number;
  onPageChange: (direction: number) => void;
  showPagination: boolean;
}

const Pagination: React.FC<PaginationProps> = ({
  currentPage,
  totalTasks,
  limit,
  onPageChange,
  showPagination
}) => {
  if (!showPagination) {
    return null;
  }

  const maxPage = Math.ceil(totalTasks / limit) - 1;
  const currentPageDisplay = currentPage + 1;
  const totalPages = maxPage + 1;

  return (
    <section className="pagination">
      <button 
        onClick={() => onPageChange(-1)}
        disabled={currentPage === 0}
      >
        Previous
      </button>
      <span id="page-info">
        Page {currentPageDisplay} of {totalPages}
      </span>
      <button 
        onClick={() => onPageChange(1)}
        disabled={currentPage >= maxPage}
      >
        Next
      </button>
    </section>
  );
};

export default Pagination;