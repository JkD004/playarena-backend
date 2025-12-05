-- phpMyAdmin SQL Dump
-- version 5.2.1
-- https://www.phpmyadmin.net/
--
-- Host: 127.0.0.1
-- Generation Time: Nov 29, 2025 at 05:32 PM
-- Server version: 10.4.32-MariaDB
-- PHP Version: 8.2.12

SET SQL_MODE = "NO_AUTO_VALUE_ON_ZERO";
START TRANSACTION;
SET time_zone = "+00:00";


/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!40101 SET NAMES utf8mb4 */;

--
-- Database: `playarena`
--

-- --------------------------------------------------------

--
-- Table structure for table `bookings`
--

CREATE TABLE `bookings` (
  `id` int(11) NOT NULL,
  `user_id` int(11) NOT NULL,
  `venue_id` int(11) NOT NULL,
  `start_time` datetime NOT NULL,
  `end_time` datetime NOT NULL,
  `total_price` decimal(10,2) NOT NULL,
  `status` enum('pending','confirmed','canceled') NOT NULL DEFAULT 'pending',
  `created_at` timestamp NOT NULL DEFAULT current_timestamp()
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

--
-- Dumping data for table `bookings`
--

INSERT INTO `bookings` (`id`, `user_id`, `venue_id`, `start_time`, `end_time`, `total_price`, `status`, `created_at`) VALUES
(3, 7, 2, '2025-10-22 14:00:00', '2025-10-22 15:00:00', 1000.00, 'confirmed', '2025-10-21 16:31:02'),
(6, 7, 3, '2025-11-07 13:00:00', '2025-11-07 14:00:00', 1200.00, 'confirmed', '2025-11-04 18:54:24'),
(7, 8, 3, '2025-11-13 12:31:00', '2025-11-13 13:31:00', 1200.00, 'confirmed', '2025-11-05 07:32:13'),
(8, 2, 4, '2025-11-20 06:30:00', '2025-11-20 07:30:00', 0.00, 'confirmed', '2025-11-19 18:18:02'),
(9, 9, 4, '2025-11-20 07:35:00', '2025-11-20 08:35:00', 400.00, 'confirmed', '2025-11-19 18:19:43'),
(10, 9, 7, '2025-11-23 18:58:00', '2025-11-23 19:58:00', 1000.00, 'canceled', '2025-11-19 18:58:08'),
(11, 9, 8, '2025-11-21 07:38:00', '2025-11-21 08:38:00', 60.00, 'confirmed', '2025-11-19 19:38:39'),
(12, 9, 8, '2025-11-22 07:45:00', '2025-11-22 08:45:00', 60.00, 'confirmed', '2025-11-19 19:45:17'),
(13, 9, 5, '2025-11-19 07:30:00', '2025-11-19 08:30:00', 1200.00, 'pending', '2025-11-19 20:43:30'),
(14, 13, 5, '2025-11-20 09:30:00', '2025-11-20 10:30:00', 1200.00, 'pending', '2025-11-20 17:50:53');

-- --------------------------------------------------------

--
-- Table structure for table `notifications`
--

CREATE TABLE `notifications` (
  `id` int(11) NOT NULL,
  `user_id` int(11) NOT NULL,
  `message` text NOT NULL,
  `type` enum('info','success','warning','error') NOT NULL DEFAULT 'info',
  `is_read` tinyint(1) NOT NULL DEFAULT 0,
  `created_at` timestamp NOT NULL DEFAULT current_timestamp()
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

--
-- Dumping data for table `notifications`
--

INSERT INTO `notifications` (`id`, `user_id`, `message`, `type`, `is_read`, `created_at`) VALUES
(1, 9, 'Payment successful! Your booking has been confirmed.', 'success', 0, '2025-11-19 19:45:46');

-- --------------------------------------------------------

--
-- Table structure for table `reviews`
--

CREATE TABLE `reviews` (
  `id` int(11) NOT NULL,
  `venue_id` int(11) NOT NULL,
  `user_id` int(11) NOT NULL,
  `rating` int(11) NOT NULL CHECK (`rating` >= 1 and `rating` <= 5),
  `comment` text DEFAULT NULL,
  `created_at` timestamp NOT NULL DEFAULT current_timestamp()
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- --------------------------------------------------------

--
-- Table structure for table `site_settings`
--

CREATE TABLE `site_settings` (
  `setting_key` varchar(50) NOT NULL,
  `setting_value` text DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

--
-- Dumping data for table `site_settings`
--

INSERT INTO `site_settings` (`setting_key`, `setting_value`) VALUES
('terms', 'Welcome to SportGrid. 1. Booking Confirmation: Your booking is confirmed only after full payment. 2. Cancellation Policy: Cancellations made less than 24 hours before the slot time are non-refundable.');

-- --------------------------------------------------------

--
-- Table structure for table `teams`
--

CREATE TABLE `teams` (
  `id` int(11) NOT NULL,
  `name` varchar(100) NOT NULL,
  `owner_id` int(11) NOT NULL,
  `created_at` timestamp NOT NULL DEFAULT current_timestamp()
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

--
-- Dumping data for table `teams`
--

INSERT INTO `teams` (`id`, `name`, `owner_id`, `created_at`) VALUES
(1, 'The All-Stars', 7, '2025-11-04 16:57:48'),
(2, 'football c div', 7, '2025-11-04 17:16:26');

-- --------------------------------------------------------

--
-- Table structure for table `team_members`
--

CREATE TABLE `team_members` (
  `id` int(11) NOT NULL,
  `team_id` int(11) NOT NULL,
  `user_id` int(11) NOT NULL,
  `status` enum('pending','joined') NOT NULL DEFAULT 'pending',
  `joined_at` timestamp NULL DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

--
-- Dumping data for table `team_members`
--

INSERT INTO `team_members` (`id`, `team_id`, `user_id`, `status`, `joined_at`) VALUES
(1, 1, 9, 'joined', NULL),
(2, 2, 7, 'joined', NULL),
(3, 1, 8, 'pending', NULL);

-- --------------------------------------------------------

--
-- Table structure for table `team_messages`
--

CREATE TABLE `team_messages` (
  `id` int(11) NOT NULL,
  `team_id` int(11) NOT NULL,
  `user_id` int(11) NOT NULL,
  `message_content` text NOT NULL,
  `created_at` timestamp NOT NULL DEFAULT current_timestamp()
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- --------------------------------------------------------

--
-- Table structure for table `users`
--

CREATE TABLE `users` (
  `id` int(11) NOT NULL,
  `first_name` varchar(100) NOT NULL,
  `last_name` varchar(100) NOT NULL,
  `phone` varchar(20) DEFAULT NULL,
  `dob` date DEFAULT NULL,
  `address` text DEFAULT NULL,
  `email` varchar(255) NOT NULL,
  `password_hash` varchar(255) NOT NULL,
  `created_at` timestamp NOT NULL DEFAULT current_timestamp(),
  `role` enum('player','owner','admin') NOT NULL DEFAULT 'player',
  `avatar_url` varchar(255) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

--
-- Dumping data for table `users`
--

INSERT INTO `users` (`id`, `first_name`, `last_name`, `phone`, `dob`, `address`, `email`, `password_hash`, `created_at`, `role`, `avatar_url`) VALUES
(1, 'Shreyas', 'jkd', '1234567890', '2018-06-12', 'Guruprasad Nagar', 'jam@gmail.com', '$2a$10$RR/TGahyq1avVeItUgzutewbnZcst8zqN9V0PcvHJbYNPow0aL8zq', '2025-10-18 16:25:27', 'admin', NULL),
(2, 'anish', 'K', '1234567891', '2025-10-19', 'Near railway station', 'anish123@gmail.com', '$2a$10$8hMB/.3Xr.uAJ8msozgXdeOcFA6cbd9FCVhiIwBriqocxvbFlgJG6', '2025-10-20 05:18:21', 'owner', NULL),
(7, 'Dummy', 'hai', '2345671890', '2022-06-07', 'Dummy Nagar', 'Hanga@gmail.com', '$2a$10$ywonY57N0Z1Cm9knhxlYMuB2LH.HM1QPI2SJChpf0a6XV0ItAKPrq', '2025-10-21 12:33:36', 'player', NULL),
(8, 'Monishya', 'K', '4567890123', '2025-11-03', 'College', 'mon@gmail.com', '$2a$10$4DclAJvEuyTTlhDMPBSC4eyJjCDBH0NMXEzV8ESN75etZPsWDyube', '2025-11-05 07:30:24', 'player', NULL),
(9, 'Khushi', '', '7259371109', '2004-05-05', 'Bailhongal', 'khushig123@gmail.com', '$2a$10$DluqYFyqdPYCmDRhOeTn4e/Kj4RfaTMqTl5KCWOXJLtR5CkTvZGLm', '2025-11-06 01:27:08', 'player', NULL),
(10, 'Sharan', 'B', '9483545939', '1967-05-15', 'Angol Behind KLE Hospital', 'cr7turfs@gmail.com', '$2a$10$VFJMb2cosZvvvZ0FBbALqeXN/1HLo5aGXJuKuGCbXxjP736bMOa1y', '2025-11-06 02:01:14', 'owner', 'https://res.cloudinary.com/dlbub74cs/image/upload/v1763579894/playarena_users/bkyjv26nopt08scr4ylj.png'),
(11, 'Aditya', 'K', '9158823893', '1967-10-18', 'Tilakwadi', 'ka22turfs@gmail.com', '$2a$10$Ti5miT7.VhBYHSgU7IjcLetnYtQK7VLd2IFOa3CzhgY52CvhjiuwO', '2025-11-06 02:18:47', 'owner', NULL),
(12, 'Sourabh', 'K', '9845772851', '1993-03-26', 'Jakkeri Honda,Near Railway Overbridge', 'sportingplanet123@gmail.com', '$2a$10$qTLwI51102UrPvJ57BOgDOZon0zAjR/eOCve/OkF8dJ2ZU5G1uIuO', '2025-11-06 05:23:15', 'owner', NULL),
(13, 'Rajesh', 'Gadad', '7947435094', '0000-00-00', '22, Ranade Colony, Near Belagavi, Khanapur Road, Tilakwadi-590006', 'orianq@gmail.com', '$2a$10$9HUi7nTg3rdj/iDRCfqbre7lwZuFN1mC0jup7iWDRhUxYKdT4BENK', '2025-11-06 05:52:12', 'owner', NULL);

-- --------------------------------------------------------

--
-- Table structure for table `venues`
--

CREATE TABLE `venues` (
  `id` int(11) NOT NULL,
  `owner_id` int(11) NOT NULL,
  `status` enum('pending','approved','rejected') NOT NULL DEFAULT 'pending',
  `name` varchar(255) NOT NULL,
  `sport_category` varchar(100) NOT NULL,
  `description` text DEFAULT NULL,
  `address` text DEFAULT NULL,
  `price_per_hour` decimal(10,2) DEFAULT NULL,
  `created_at` timestamp NOT NULL DEFAULT current_timestamp(),
  `opening_time` varchar(10) NOT NULL DEFAULT '06:00',
  `closing_time` varchar(10) NOT NULL DEFAULT '23:00',
  `lunch_start_time` varchar(10) DEFAULT NULL,
  `lunch_end_time` varchar(10) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

--
-- Dumping data for table `venues`
--

INSERT INTO `venues` (`id`, `owner_id`, `status`, `name`, `sport_category`, `description`, `address`, `price_per_hour`, `created_at`, `opening_time`, `closing_time`, `lunch_start_time`, `lunch_end_time`) VALUES
(2, 2, 'approved', 'crucibles', 'Snooker', 'Get the best experience of snooker here ', 'Near Rockey Cha point', 60.00, '2025-10-21 14:56:16', '06:00', '23:00', NULL, NULL),
(3, 2, 'approved', 'Lotus turf', 'Football', 'come and play', 'Near lotus county', 1200.00, '2025-10-21 15:23:23', '06:00', '23:00', NULL, NULL),
(4, 2, 'approved', 'KLE', 'Badminton', 'BEST ', 'Udyambag', 400.00, '2025-11-05 07:27:12', '06:00', '23:00', NULL, NULL),
(5, 10, 'approved', 'CR7 Sports Arena', 'Football', 'Belgaum\'s  largest turf ground', 'Yellur Road, Behind Kle Hospital', 1200.00, '2025-11-06 02:09:16', '10:00', '23:00', '12:30', '13:45'),
(6, 11, 'approved', 'Turf KA22', 'Football', 'The Biggest turf in Belgaum which is Available for football and Box cricket', 'Balika Adarsh Vidyalaya Ground Near Ram mandir', 800.00, '2025-11-06 02:21:27', '06:00', '23:00', NULL, NULL),
(7, 12, 'approved', 'Sporting Planet Sports Complex', 'Football', 'Multi sport turf facility located inside Blooming Buds School, ideal for football, cricket, and recreational matches.', 'Jakkeri Honda', 1000.00, '2025-11-06 05:29:48', '06:00', '23:00', NULL, NULL),
(8, 13, 'approved', 'Orian Q', 'Snooker', 'Choose among Multiple table options, and open from morning 11 to evening 9', 'RPD', 60.00, '2025-11-06 05:56:33', '06:00', '23:00', NULL, NULL);

-- --------------------------------------------------------

--
-- Table structure for table `venue_photos`
--

CREATE TABLE `venue_photos` (
  `id` int(11) NOT NULL,
  `venue_id` int(11) NOT NULL,
  `image_url` varchar(255) NOT NULL,
  `created_at` timestamp NOT NULL DEFAULT current_timestamp()
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

--
-- Dumping data for table `venue_photos`
--

INSERT INTO `venue_photos` (`id`, `venue_id`, `image_url`, `created_at`) VALUES
(1, 6, 'https://res.cloudinary.com/dlbub74cs/image/upload/v1762396477/playarena_venues/cywzpfiwfn9mpwdraijq.webp', '2025-11-06 02:34:37');

--
-- Indexes for dumped tables
--

--
-- Indexes for table `bookings`
--
ALTER TABLE `bookings`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `unique_booking` (`venue_id`,`start_time`),
  ADD KEY `user_id` (`user_id`);

--
-- Indexes for table `notifications`
--
ALTER TABLE `notifications`
  ADD PRIMARY KEY (`id`),
  ADD KEY `user_id` (`user_id`);

--
-- Indexes for table `reviews`
--
ALTER TABLE `reviews`
  ADD PRIMARY KEY (`id`),
  ADD KEY `venue_id` (`venue_id`),
  ADD KEY `user_id` (`user_id`);

--
-- Indexes for table `site_settings`
--
ALTER TABLE `site_settings`
  ADD PRIMARY KEY (`setting_key`);

--
-- Indexes for table `teams`
--
ALTER TABLE `teams`
  ADD PRIMARY KEY (`id`),
  ADD KEY `owner_id` (`owner_id`);

--
-- Indexes for table `team_members`
--
ALTER TABLE `team_members`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `unique_member` (`team_id`,`user_id`),
  ADD KEY `user_id` (`user_id`);

--
-- Indexes for table `team_messages`
--
ALTER TABLE `team_messages`
  ADD PRIMARY KEY (`id`),
  ADD KEY `team_id` (`team_id`),
  ADD KEY `user_id` (`user_id`);

--
-- Indexes for table `users`
--
ALTER TABLE `users`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `email` (`email`),
  ADD UNIQUE KEY `uc_email` (`email`),
  ADD UNIQUE KEY `uc_phone` (`phone`);

--
-- Indexes for table `venues`
--
ALTER TABLE `venues`
  ADD PRIMARY KEY (`id`),
  ADD KEY `owner_id` (`owner_id`);

--
-- Indexes for table `venue_photos`
--
ALTER TABLE `venue_photos`
  ADD PRIMARY KEY (`id`),
  ADD KEY `venue_id` (`venue_id`);

--
-- AUTO_INCREMENT for dumped tables
--

--
-- AUTO_INCREMENT for table `bookings`
--
ALTER TABLE `bookings`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=15;

--
-- AUTO_INCREMENT for table `notifications`
--
ALTER TABLE `notifications`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=2;

--
-- AUTO_INCREMENT for table `reviews`
--
ALTER TABLE `reviews`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT for table `teams`
--
ALTER TABLE `teams`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=3;

--
-- AUTO_INCREMENT for table `team_members`
--
ALTER TABLE `team_members`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=4;

--
-- AUTO_INCREMENT for table `team_messages`
--
ALTER TABLE `team_messages`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT for table `users`
--
ALTER TABLE `users`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=14;

--
-- AUTO_INCREMENT for table `venues`
--
ALTER TABLE `venues`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=9;

--
-- AUTO_INCREMENT for table `venue_photos`
--
ALTER TABLE `venue_photos`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=2;

--
-- Constraints for dumped tables
--

--
-- Constraints for table `bookings`
--
ALTER TABLE `bookings`
  ADD CONSTRAINT `bookings_ibfk_1` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`),
  ADD CONSTRAINT `bookings_ibfk_2` FOREIGN KEY (`venue_id`) REFERENCES `venues` (`id`);

--
-- Constraints for table `notifications`
--
ALTER TABLE `notifications`
  ADD CONSTRAINT `notifications_ibfk_1` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE;

--
-- Constraints for table `reviews`
--
ALTER TABLE `reviews`
  ADD CONSTRAINT `reviews_ibfk_1` FOREIGN KEY (`venue_id`) REFERENCES `venues` (`id`) ON DELETE CASCADE,
  ADD CONSTRAINT `reviews_ibfk_2` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`);

--
-- Constraints for table `teams`
--
ALTER TABLE `teams`
  ADD CONSTRAINT `teams_ibfk_1` FOREIGN KEY (`owner_id`) REFERENCES `users` (`id`) ON DELETE CASCADE;

--
-- Constraints for table `team_members`
--
ALTER TABLE `team_members`
  ADD CONSTRAINT `team_members_ibfk_1` FOREIGN KEY (`team_id`) REFERENCES `teams` (`id`) ON DELETE CASCADE,
  ADD CONSTRAINT `team_members_ibfk_2` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE;

--
-- Constraints for table `team_messages`
--
ALTER TABLE `team_messages`
  ADD CONSTRAINT `team_messages_ibfk_1` FOREIGN KEY (`team_id`) REFERENCES `teams` (`id`) ON DELETE CASCADE,
  ADD CONSTRAINT `team_messages_ibfk_2` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE;

--
-- Constraints for table `venues`
--
ALTER TABLE `venues`
  ADD CONSTRAINT `venues_ibfk_1` FOREIGN KEY (`owner_id`) REFERENCES `users` (`id`);

--
-- Constraints for table `venue_photos`
--
ALTER TABLE `venue_photos`
  ADD CONSTRAINT `venue_photos_ibfk_1` FOREIGN KEY (`venue_id`) REFERENCES `venues` (`id`) ON DELETE CASCADE;
COMMIT;

/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;