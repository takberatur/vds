package com.agcforge.videodownloader

import android.annotation.SuppressLint
import android.app.Application
import android.content.Context
import com.agcforge.videodownloader.utils.DownloadManagerCleaner
import com.onesignal.OneSignal
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.SupervisorJob

import kotlinx.coroutines.launch



class App : Application() {

    private val TAG = "App_Root_Main"

    companion object {
        @SuppressLint("StaticFieldLeak")
        @Volatile
        private var instance: App? = null

        fun getInstance(): App {
            return instance ?: throw IllegalStateException("App not initialized")
        }

        fun getContext(): Context {
            return getInstance().applicationContext
        }
    }

    private val applicationScope = CoroutineScope(SupervisorJob() + Dispatchers.Main)

    override fun onCreate() {
        super.onCreate()
        instance = this
        DownloadManagerCleaner.clearFailedDownloads(this)

    }
}
