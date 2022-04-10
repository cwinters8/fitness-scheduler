-- Note: Using a Planetscale database, which means foreign keys are not possible.
-- The application needs to handle relationships between tables.

drop table `sessions`;
drop table routines;
drop table frequencies;
drop table frequency_days;
drop table reminders;

create table `sessions` (
  id int not null auto_increment,
  primary key (id),
  `user_id` int not null,
  title text not null,
  routine_id int not null,
  `timestamp` timestamp not null,
  duration int not null,
  frequency_id int not null,
  notes text null
);

create table routines (
  id int not null auto_increment,
  primary key (id),
  `name` text not null,
  category text not null,
  `description` text not null,
  `url` text not null,
  duration int not null,
  votes int not null,
  `user_id` int not null,
  public boolean not null,
  views int not null,
  times_completed int not null,
  created timestamp not null,
  modified timestamp null
);

create table frequencies (
  id int not null auto_increment,
  primary key (id),
  `start_date` date null,
  `end_date` date null,
  `type` text not null
);

create table frequency_days (
  id int not null auto_increment,
  primary key (id),
  frequency_id int not null,
  `day` int not null
);

create table reminders (
  id int not null auto_increment,
  primary key (id),
  `session_id` int not null,
  `time` timestamp null,
  minutes_prior int null,
  status text not null default 'pending'
);
