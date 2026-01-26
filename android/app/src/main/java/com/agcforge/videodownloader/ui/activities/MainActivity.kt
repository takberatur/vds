package com.agcforge.videodownloader.ui.activities

import android.content.Intent
import android.os.Bundle
import android.view.MenuItem
import androidx.activity.OnBackPressedCallback
import androidx.core.view.GravityCompat
import androidx.lifecycle.lifecycleScope
import androidx.navigation.NavController
import androidx.navigation.fragment.NavHostFragment
import androidx.navigation.ui.AppBarConfiguration
import androidx.navigation.ui.navigateUp
import androidx.navigation.ui.setupActionBarWithNavController
import androidx.navigation.ui.setupWithNavController
import com.agcforge.videodownloader.R
import com.agcforge.videodownloader.databinding.ActivityMainBinding
import com.agcforge.videodownloader.service.WebSocketService
import com.agcforge.videodownloader.ui.activities.auth.LoginActivity
import com.agcforge.videodownloader.utils.PreferenceManager
import com.agcforge.videodownloader.utils.loadImage
import com.google.android.material.dialog.MaterialAlertDialogBuilder
import com.google.android.material.navigation.NavigationView
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.launch

class MainActivity : BaseActivity(), NavigationView.OnNavigationItemSelectedListener {

    private lateinit var binding: ActivityMainBinding
    private lateinit var navController: NavController
    private lateinit var appBarConfiguration: AppBarConfiguration
    private lateinit var preferenceManager: PreferenceManager

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        binding = ActivityMainBinding.inflate(layoutInflater)
        setContentView(binding.root)

        preferenceManager = PreferenceManager(this)

        setupToolbar()
        setupNavigation()
        setupDrawer()
        updateNavigationHeader()
        startWebSocketService()
    }

    private fun setupToolbar() {
        setSupportActionBar(binding.toolbar)
    }

    private fun setupNavigation() {
        val navHostFragment = supportFragmentManager
            .findFragmentById(R.id.navHostFragment) as NavHostFragment
        navController = navHostFragment.navController

        // Setup bottom navigation
        binding.bottomNavigation.setupWithNavController(navController)

        // Setup app bar configuration
        appBarConfiguration = AppBarConfiguration(
            setOf(
                R.id.homeFragment,
                R.id.downloadsFragment,
                R.id.settingsFragment
            ),
            binding.drawerLayout
        )

        setupActionBarWithNavController(navController, appBarConfiguration)
    }

    private fun setupDrawer() {
        val toggle = androidx.appcompat.app.ActionBarDrawerToggle(
            this,
            binding.drawerLayout,
            binding.toolbar,
            R.string.navigation_drawer_open,
            R.string.navigation_drawer_close
        )
        binding.drawerLayout.addDrawerListener(toggle)
        toggle.syncState()

        binding.navigationView.setNavigationItemSelectedListener(this)
    }

    private fun updateNavigationHeader() {
        val headerView = binding.navigationView.getHeaderView(0)
        val ivAvatar = headerView.findViewById<android.widget.ImageView>(R.id.ivAvatar)
        val tvUserName = headerView.findViewById<android.widget.TextView>(R.id.tvUserName)
        val tvUserEmail = headerView.findViewById<android.widget.TextView>(R.id.tvUserEmail)

        lifecycleScope.launch {
            val userName = preferenceManager.userId.first() // You can store user name separately
            val userEmail = preferenceManager.authToken.first() // Get actual email from preferences

            tvUserName.text = userName ?: "Guest User"
            tvUserEmail.text = userEmail ?: "guest@example.com"

            // Load avatar if available
            // ivAvatar.loadImage(avatarUrl)
        }
    }

    private fun startWebSocketService() {
        lifecycleScope.launch {
            val userId = preferenceManager.userId.first()
            val token = preferenceManager.authToken.first()

            if (!userId.isNullOrEmpty() && !token.isNullOrEmpty()) {
                WebSocketService.start(this@MainActivity, userId, token)
            }
        }
    }

    override fun onNavigationItemSelected(item: MenuItem): Boolean {
        when (item.itemId) {
            R.id.nav_home -> navController.navigate(R.id.homeFragment)
            R.id.nav_downloads -> navController.navigate(R.id.downloadsFragment)
            R.id.nav_settings -> navController.navigate(R.id.settingsFragment)
            R.id.nav_about -> {
                showAboutDialog()
            }
            R.id.nav_logout -> {
                showLogoutDialog()
            }
        }
        binding.drawerLayout.closeDrawer(GravityCompat.START)
        return true
    }

    private fun showAboutDialog() {
        MaterialAlertDialogBuilder(this)
            .setTitle(R.string.about)
            .setMessage(R.string.about_dialog_message)
            .setPositiveButton(R.string.ok, null)
            .show()
    }

    private fun showLogoutDialog() {
        MaterialAlertDialogBuilder(this)
            .setTitle(R.string.logout)
            .setMessage("Are you sure you want to logout?")
            .setPositiveButton(R.string.logout) { _, _ ->
                handleLogout()
            }
            .setNegativeButton("Cancel", null)
            .show()
    }

    private fun handleLogout() {
        lifecycleScope.launch {
            // Stop WebSocket service
            WebSocketService.stop(this@MainActivity)

            // Clear user data
            preferenceManager.clearUserData()

            // Navigate to login
            startActivity(Intent(this@MainActivity, LoginActivity::class.java).apply {
                flags = Intent.FLAG_ACTIVITY_NEW_TASK or Intent.FLAG_ACTIVITY_CLEAR_TASK
            })
            finish()
        }
    }

    override fun onSupportNavigateUp(): Boolean {
        return navController.navigateUp(appBarConfiguration) || super.onSupportNavigateUp()
    }

    private fun setupBackPressedCallback() {
        val callback = object : OnBackPressedCallback(true) {
            override fun handleOnBackPressed() {
                if (binding.drawerLayout.isDrawerOpen(GravityCompat.START)) {
                    binding.drawerLayout.closeDrawer(GravityCompat.START)
                } else {
                    onBackPressedDispatcher.onBackPressed()
                }
            }
        }
        onBackPressedDispatcher.addCallback(this, callback)
    }
}
